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

func (s *DecoratedInventory) UpdateLabstations(ctx context.Context, req *UpdateLabstationsRequest) (rsp *UpdateLabstationsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "UpdateLabstations", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.UpdateLabstations(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "UpdateLabstations", rsp, err)
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

func (s *DecoratedInventory) BatchUpdateDevices(ctx context.Context, req *BatchUpdateDevicesRequest) (rsp *BatchUpdateDevicesResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "BatchUpdateDevices", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.BatchUpdateDevices(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "BatchUpdateDevices", rsp, err)
	}
	return
}

func (s *DecoratedInventory) AddAssets(ctx context.Context, req *AssetList) (rsp *AssetResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "AddAssets", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.AddAssets(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "AddAssets", rsp, err)
	}
	return
}

func (s *DecoratedInventory) GetAssets(ctx context.Context, req *AssetIDList) (rsp *AssetResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "GetAssets", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.GetAssets(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "GetAssets", rsp, err)
	}
	return
}

func (s *DecoratedInventory) DeleteAssets(ctx context.Context, req *AssetIDList) (rsp *AssetIDResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "DeleteAssets", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.DeleteAssets(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "DeleteAssets", rsp, err)
	}
	return
}

func (s *DecoratedInventory) UpdateAssets(ctx context.Context, req *AssetList) (rsp *AssetResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "UpdateAssets", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.UpdateAssets(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "UpdateAssets", rsp, err)
	}
	return
}

func (s *DecoratedInventory) DeviceConfigsExists(ctx context.Context, req *DeviceConfigsExistsRequest) (rsp *DeviceConfigsExistsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "DeviceConfigsExists", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.DeviceConfigsExists(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "DeviceConfigsExists", rsp, err)
	}
	return
}

func (s *DecoratedInventory) GetDeviceManualRepairRecords(ctx context.Context, req *GetDeviceManualRepairRecordsRequest) (rsp *GetDeviceManualRepairRecordsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "GetDeviceManualRepairRecords", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.GetDeviceManualRepairRecords(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "GetDeviceManualRepairRecords", rsp, err)
	}
	return
}

func (s *DecoratedInventory) CreateDeviceManualRepairRecords(ctx context.Context, req *CreateDeviceManualRepairRecordsRequest) (rsp *CreateDeviceManualRepairRecordsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "CreateDeviceManualRepairRecords", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.CreateDeviceManualRepairRecords(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "CreateDeviceManualRepairRecords", rsp, err)
	}
	return
}

func (s *DecoratedInventory) UpdateDeviceManualRepairRecords(ctx context.Context, req *UpdateDeviceManualRepairRecordsRequest) (rsp *UpdateDeviceManualRepairRecordsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "UpdateDeviceManualRepairRecords", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.UpdateDeviceManualRepairRecords(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "UpdateDeviceManualRepairRecords", rsp, err)
	}
	return
}

func (s *DecoratedInventory) ListCrosDevicesLabConfig(ctx context.Context, req *ListCrosDevicesLabConfigRequest) (rsp *ListCrosDevicesLabConfigResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "ListCrosDevicesLabConfig", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ListCrosDevicesLabConfig(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "ListCrosDevicesLabConfig", rsp, err)
	}
	return
}
