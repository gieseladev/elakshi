package wamp

import (
	"context"
	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/getsentry/sentry-go"
	"github.com/gieseladev/elakshi/pkg/api"
	"github.com/gieseladev/elakshi/pkg/errutil"
	"github.com/gieseladev/elakshi/pkg/infoextract"
	"log"
)

type wampHandler struct {
	core *api.Core
	c    *client.Client
}

// NewWAMPHandler creates a new api.Handler which provides interaction over WAMP.
// ctx must contain an API core, otherwise the function panics.
func NewWAMPHandler(ctx context.Context, c *client.Client) *wampHandler {
	core := api.CoreFromContext(ctx)
	if core == nil {
		panic("api/wamp: passed context without api core")
	}

	return &wampHandler{
		core: core,
		c:    c,
	}
}

func (s *wampHandler) registerProcedures() error {
	const elakshiPrefix = "io.elakshi."

	return errutil.CollectErrors(
		s.c.Register(elakshiPrefix+"meta.assert_ready", s.metaAssertReady, nil),

		s.c.Register(elakshiPrefix+"get_track", s.get, nil),
		s.c.Register(elakshiPrefix+"resolve", s.resolve, nil),
		s.c.Register(elakshiPrefix+"get_audio_source", s.getAudio, nil),
	)
}

func (s *wampHandler) Start() error {
	return s.registerProcedures()
}

func (s *wampHandler) Stop() error {
	return s.c.Close()
}

func (s *wampHandler) Done() <-chan struct{} {
	return s.c.Done()
}

func handleError(err error) client.InvokeResult {
	switch err {
	case api.ErrEIDNotFound:
		return InvalidArgumentResult("eid not found")
	}

	log.Println("api/wamp: unexpected error:", err)

	res := ErrorResult(ErrInternalError)

	eventID := sentry.CaptureException(err)
	if eventID != nil {
		ErrorAddKwarg(&res, "event_id", eventID)
	}

	return res
}

func (s *wampHandler) metaAssertReady(ctx context.Context, invocation *wamp.Invocation) client.InvokeResult {
	return client.InvokeResult{}
}

func (s *wampHandler) get(ctx context.Context, invocation *wamp.Invocation) client.InvokeResult {
	eid, ok := GetStrArg(invocation.Arguments, 0)
	if !ok {
		return InvalidArgumentResult("EID missing")
	}

	track, err := s.core.GetTrack(eid)
	if err != nil {
		return handleError(err)
	}

	return SingleValueResult(track)
}

func (s *wampHandler) resolve(ctx context.Context, invocation *wamp.Invocation) client.InvokeResult {
	uri, ok := GetStrArg(invocation.Arguments, 0)
	if !ok {
		return InvalidArgumentResult("uri missing")
	}

	res, err := s.core.ResolveURI(ctx, uri)
	switch err {
	case api.ErrNoExtractorForURI:
		return InvalidArgumentResult("no extractor for uri")
	case infoextract.ErrURIInvalid:
		return InvalidArgumentResult("uri invalid")
	case nil:
	default:
		return handleError(err)
	}

	resp, err := api.CreateResolveResponse(res)
	if err != nil {
		return handleError(err)
	}

	return SingleValueResult(resp)
}

func (s *wampHandler) getAudio(ctx context.Context, invocation *wamp.Invocation) client.InvokeResult {
	eid, ok := GetStrArg(invocation.Arguments, 0)
	if !ok {
		return InvalidArgumentResult("EID missing")
	}

	source, err := s.core.GetTrackSource(ctx, eid)
	if err != nil {
		return handleError(err)
	}

	return SingleValueResult(source)
}
