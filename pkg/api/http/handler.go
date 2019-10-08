package http

import (
	"context"
	"encoding/json"
	"errors"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gieseladev/elakshi/pkg/api"
	"github.com/gieseladev/elakshi/pkg/edb"
	"log"
	"net/http"
)

type httpHandler struct {
	ctx  context.Context
	core *api.Core

	srv *http.Server
	mux *http.ServeMux

	sentryHandler *sentryhttp.Handler

	done    chan struct{}
	started bool
}

func NewHTTPHandler(ctx context.Context, addr string) *httpHandler {
	core := api.CoreFromContext(ctx)
	if core == nil {
		panic("api/http: passed context without api core")
	}

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	return &httpHandler{
		ctx:  ctx,
		core: core,
		srv:  srv, mux: mux,
		sentryHandler: sentryHandler,
		done:          make(chan struct{}),
	}
}

func (h *httpHandler) Start() error {
	if h.started {
		return errors.New("already started")
	}

	h.started = true
	h.addRoutes()

	go func() {
		// when the server stopped serving, send notification by closing the
		// done channel
		err := h.srv.ListenAndServe()
		if err != nil {
			log.Println("http server stopped", err)
		}

		close(h.done)
	}()

	go func() {
		// when the context is cancelled, stop the server
		<-h.ctx.Done()
		_ = h.Stop()
	}()

	return nil
}

func (h *httpHandler) Stop() error {
	return h.srv.Shutdown(h.ctx)
}

func (h *httpHandler) Done() <-chan struct{} {
	return h.done
}

const (
	trackPath  = "/track/"
	lyricsPath = "/lyrics/"
)

func (h *httpHandler) addRoutes() {
	s := h.sentryHandler

	h.mux.HandleFunc(trackPath, s.HandleFunc(h.getTrack))
	h.mux.HandleFunc(lyricsPath, s.HandleFunc(h.getLyrics))
}

// writeJSONResponse writes json encoded data into the http response writer and
// sets the appropriate Content-Type header.
func writeJSONResponse(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

func handleError(w http.ResponseWriter, err error) {
	statusCode := 500
	switch err {
	case edb.ErrEIDInvalid:
		statusCode = 400
	case api.ErrEIDNotFound:
		statusCode = 404
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(err.Error()))
}

func (h *httpHandler) getTrack(w http.ResponseWriter, r *http.Request) {
	eid := r.URL.Path[len(trackPath):]
	track, err := h.core.GetTrack(eid)
	if err != nil {
		handleError(w, err)
		return
	}

	if err := writeJSONResponse(w, track); err != nil {
		panic(err)
	}
}

func (h *httpHandler) getLyrics(w http.ResponseWriter, r *http.Request) {
	eid := r.URL.Path[len(lyricsPath):]
	lyrics, err := h.core.GetTrackLyrics(r.Context(), eid)
	if err != nil {
		handleError(w, err)
		return
	}

	if err := writeJSONResponse(w, lyrics); err != nil {
		panic(err)
	}
}
