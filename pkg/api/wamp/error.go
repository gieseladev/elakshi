package wamp

import (
	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
)

const (
	ErrInternalError = wamp.URI("io.elakshi.internal_error")
)

func ErrorResult(uri wamp.URI) client.InvokeResult {
	return client.InvokeResult{
		Err: uri,
	}
}

func ErrorAddKwarg(res *client.InvokeResult, key string, value interface{}) {
	if res.Kwargs == nil {
		res.Kwargs = wamp.Dict{key: value}
		return
	}

	res.Kwargs[key] = value
}

func InvalidArgumentResult(message string) client.InvokeResult {
	res := ErrorResult(wamp.ErrInvalidArgument)
	res.Args = append(res.Args, message)

	return res
}

func EIDMissingResult() client.InvokeResult {
	return InvalidArgumentResult("EID missing")
}
