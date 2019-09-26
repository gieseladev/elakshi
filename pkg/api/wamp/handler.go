package wamp

//import (
//	"context"
//	"github.com/gammazero/nexus/client"
//	"github.com/gammazero/nexus/wamp"
//	"github.com/gieseladev/elakshi/pkg/errutils"
//)
//
//type wampHandler struct {
//	core *Core
//	c *client.Client
//}
//
//func NewWAMPHandler(c *client.Client) *wampHandler {
//	return &wampHandler{
//		c: c,
//	}
//}
//
//func (s *wampHandler) registerProcedures() error {
//	return errutils.CollectErrors(
//		s.c.Register("io.giesela.elakshi.get", s.get, wamp.Dict{}),
//	)
//}
//
//func (s *wampHandler) Start() error {
//	return s.registerProcedures()
//}
//
//func (s *wampHandler) Stop() error {
//	return s.c.Close()
//}
//
//func (s *wampHandler) Done() <-chan struct{} {
//	return s.c.Done()
//}
//
//func (s *wampHandler) get(ctx context.Context, args wamp.List, kwargs wamp.Dict, options wamp.Dict) *client.InvokeResult {
//	return &client.InvokeResult{}
//}
