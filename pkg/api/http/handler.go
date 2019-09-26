package http

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gieseladev/elakshi/pkg/api"
	"net/http"
)

type httpHandler struct {
	ctx  context.Context
	core *api.Core

	srv *http.Server
	mux *http.ServeMux

	done    chan struct{}
	started bool
}

func NewHTTPHandler(ctx context.Context, addr string) *httpHandler {
	core := api.CoreFromContext(ctx)
	if core == nil {
		panic("context without core passed")
	}

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &httpHandler{
		ctx:  ctx,
		core: core,
		srv:  srv, mux: mux,
		done: make(chan struct{}),
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
		_ = h.srv.ListenAndServe()
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
	trackPath    = "/track/"
	trackPathLen = len(trackPath)
)

func (h *httpHandler) addRoutes() {
	h.mux.HandleFunc(trackPath, h.getTrack)
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
	case api.ErrEIDInvalid:
		statusCode = 400
	case api.ErrEIDNotFound:
		statusCode = 404
	}

	w.WriteHeader(statusCode)
}

func (h *httpHandler) getTrack(w http.ResponseWriter, r *http.Request) {
	eid := r.URL.Path[trackPathLen:]
	track, err := api.GetTrack(h.core.DB, eid)
	if err != nil {
		handleError(w, err)
		return
	}

	if err := writeJSONResponse(w, track); err != nil {
		panic(err)
	}
}
