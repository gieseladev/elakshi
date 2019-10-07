package wamp

import (
	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
)

func GetArg(args wamp.List, i int) (interface{}, bool) {
	if !(i >= 0 && i < len(args)) {
		return nil, false
	}

	return args[i], true
}

func GetEID(args wamp.List, i int) (string, bool) {
	if arg, ok := GetArg(args, i); ok {
		return wamp.AsString(arg)
	}

	return "", false
}

func SingleValueResult(value interface{}) client.InvokeResult {
	return client.InvokeResult{
		Args: wamp.List{value},
	}
}
