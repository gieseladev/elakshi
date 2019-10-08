package wamp

import (
	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
)

const (
	// ErrInternalError represents an internal error.
	ErrInternalError = wamp.URI("io.elakshi.internal_error")
)

// ErrorResult creates a new client.InvokeResult with the given error uri.
func ErrorResult(uri wamp.URI) client.InvokeResult {
	return client.InvokeResult{
		Err: uri,
	}
}

// ErrorAddKwarg adds a keyword argument to the given client.InvokeResult.
func ErrorAddKwarg(res *client.InvokeResult, key string, value interface{}) {
	if res.Kwargs == nil {
		res.Kwargs = wamp.Dict{key: value}
		return
	}

	res.Kwargs[key] = value
}

// InvalidArgumentResult creates a new client.InvokeResult with the
// ErrInvalidArgument error uri and the given message as its first argument.
func InvalidArgumentResult(message string) client.InvokeResult {
	res := ErrorResult(wamp.ErrInvalidArgument)
	res.Args = append(res.Args, message)

	return res
}
