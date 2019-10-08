package wamp

import (
	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
)

// GetArg returns the ith item from a list.
func GetArg(args wamp.List, i int) (interface{}, bool) {
	if !(i >= 0 && i < len(args)) {
		return nil, false
	}

	return args[i], true
}

// GetStrArg returns the ith item from a list as a string.
func GetStrArg(args wamp.List, i int) (string, bool) {
	if arg, ok := GetArg(args, i); ok {
		return wamp.AsString(arg)
	}

	return "", false
}

// SingleValueResult returns a wamp client.InvokeResult with the given value as
// its only argument.
func SingleValueResult(value interface{}) client.InvokeResult {
	return client.InvokeResult{
		Args: wamp.List{value},
	}
}
