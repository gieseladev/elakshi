package wamp

import (
	"context"
	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/getsentry/sentry-go"
	"github.com/gieseladev/elakshi/pkg/api"
	"github.com/gieseladev/elakshi/pkg/errutil"
	"log"
)

type wampHandler struct {
	core *api.Core
	c    *client.Client
}

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
	return errutil.CollectErrors(
		s.c.Register("io.giesela.elakshi.get", s.get, wamp.Dict{}),
		s.c.Register("io.giesela.elakshi.resolve", s.resolve, wamp.Dict{}),
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

func (s *wampHandler) get(ctx context.Context, invocation *wamp.Invocation) client.InvokeResult {
	eid, ok := GetEID(invocation.Arguments, 0)
	if !ok {
		return EIDMissingResult()
	}

	track, err := s.core.GetTrack(eid)
	if err != nil {
		return handleError(err)
	}

	return SingleValueResult(track)
}

func (s *wampHandler) resolve(ctx context.Context, invocation *wamp.Invocation) client.InvokeResult {
	return client.InvokeResult{}
}
