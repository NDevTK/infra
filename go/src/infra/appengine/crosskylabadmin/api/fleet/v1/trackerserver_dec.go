// Code generated by svcdec; DO NOT EDIT.

package fleet

import (
	"context"

	proto "github.com/golang/protobuf/proto"
)

type DecoratedTracker struct {
	// Service is the service to decorate.
	Service TrackerServer
	// Prelude is called for each method before forwarding the call to Service.
	// If Prelude returns an error, then the call is skipped and the error is
	// processed via the Postlude (if one is defined), or it is returned directly.
	Prelude func(c context.Context, methodName string, req proto.Message) (context.Context, error)
	// Postlude is called for each method after Service has processed the call, or
	// after the Prelude has returned an error. This takes the the Service's
	// response proto (which may be nil) and/or any error. The decorated
	// service will return the response (possibly mutated) and error that Postlude
	// returns.
	Postlude func(c context.Context, methodName string, rsp proto.Message, err error) error
}

func (s *DecoratedTracker) RefreshBots(c context.Context, req *RefreshBotsRequest) (rsp *RefreshBotsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(c, "RefreshBots", req)
		if err == nil {
			c = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.RefreshBots(c, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(c, "RefreshBots", rsp, err)
	}
	return
}

func (s *DecoratedTracker) SummarizeBots(c context.Context, req *SummarizeBotsRequest) (rsp *SummarizeBotsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(c, "SummarizeBots", req)
		if err == nil {
			c = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.SummarizeBots(c, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(c, "SummarizeBots", rsp, err)
	}
	return
}

func (s *DecoratedTracker) PushBotsForAdminTasks(c context.Context, req *PushBotsForAdminTasksRequest) (rsp *PushBotsForAdminTasksResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(c, "PushBotsForAdminTasks", req)
		if err == nil {
			c = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.PushBotsForAdminTasks(c, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(c, "PushBotsForAdminTasks", rsp, err)
	}
	return
}

func (s *DecoratedTracker) ReportBots(c context.Context, req *ReportBotsRequest) (rsp *ReportBotsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(c, "ReportBots", req)
		if err == nil {
			c = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ReportBots(c, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(c, "ReportBots", rsp, err)
	}
	return
}
