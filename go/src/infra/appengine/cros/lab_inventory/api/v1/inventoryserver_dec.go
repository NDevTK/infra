// Code generated by svcdec; DO NOT EDIT.

package api

import (
	"context"

	proto "github.com/golang/protobuf/proto"
)

type DecoratedInventory struct {
	// Service is the service to decorate.
	Service InventoryServer
	// Prelude is called for each method before forwarding the call to Service.
	// If Prelude returns an error, then the call is skipped and the error is
	// processed via the Postlude (if one is defined), or it is returned directly.
	Prelude func(ctx context.Context, methodName string, req proto.Message) (context.Context, error)
	// Postlude is called for each method after Service has processed the call, or
	// after the Prelude has returned an error. This takes the the Service's
	// response proto (which may be nil) and/or any error. The decorated
	// service will return the response (possibly mutated) and error that Postlude
	// returns.
	Postlude func(ctx context.Context, methodName string, rsp proto.Message, err error) error
}

func (s *DecoratedInventory) AddCrosDevices(ctx context.Context, req *AddCrosDevicesRequest) (rsp *AddCrosDevicesResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "AddCrosDevices", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.AddCrosDevices(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "AddCrosDevices", rsp, err)
	}
	return
}

func (s *DecoratedInventory) GetCrosDevices(ctx context.Context, req *GetCrosDevicesRequest) (rsp *GetCrosDevicesResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "GetCrosDevices", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.GetCrosDevices(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "GetCrosDevices", rsp, err)
	}
	return
}

func (s *DecoratedInventory) UpdateDutsStatus(ctx context.Context, req *UpdateDutsStatusRequest) (rsp *UpdateDutsStatusResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "UpdateDutsStatus", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.UpdateDutsStatus(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "UpdateDutsStatus", rsp, err)
	}
	return
}

func (s *DecoratedInventory) UpdateCrosDevicesSetup(ctx context.Context, req *UpdateCrosDevicesSetupRequest) (rsp *UpdateCrosDevicesSetupResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "UpdateCrosDevicesSetup", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.UpdateCrosDevicesSetup(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "UpdateCrosDevicesSetup", rsp, err)
	}
	return
}

func (s *DecoratedInventory) DeleteCrosDevices(ctx context.Context, req *DeleteCrosDevicesRequest) (rsp *DeleteCrosDevicesResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "DeleteCrosDevices", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.DeleteCrosDevices(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "DeleteCrosDevices", rsp, err)
	}
	return
}
