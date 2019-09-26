/*
Package api defines the apis supported by Elakshi.
*/

package api

// Handler defines the interface for an API request handler.
type Handler interface {
	// Start starts the handler. The handler should start responding to
	// requests after being started.
	Start() error

	// Stop closes the handler. The handler should no longer respond to
	// requests after being closed.
	Stop() error

	// Done returns a channel which is closed when the handler is done.
	Done() <-chan struct{}
}

type handlerSlice []Handler

// CollectHandlers creates a new handler from the given handlers.
func CollectHandlers(handlers ...Handler) Handler {
	if len(handlers) == 1 {
		return handlers[0]
	}

	return handlerSlice(handlers)
}

func (h handlerSlice) forEach(f func(Handler) error) error {
	for _, handler := range h {
		if err := f(handler); err != nil {
			return err
		}
	}

	return nil
}

func (h handlerSlice) Start() error {
	return h.forEach(func(handler Handler) error { return handler.Start() })
}

func (h handlerSlice) Stop() error {
	return h.forEach(func(handler Handler) error { return handler.Stop() })
}

func (h handlerSlice) Done() <-chan struct{} {
	done := make(chan struct{}, 0)

	go func() {
		defer close(done)

		for _, handler := range h {
			<-handler.Done()
		}
	}()

	return done
}
