// Code generated by svcdec; DO NOT EDIT.

package ufspb

import (
	"context"

	proto "github.com/golang/protobuf/proto"

	empty "github.com/golang/protobuf/ptypes/empty"
	status "google.golang.org/genproto/googleapis/rpc/status"
	proto1 "infra/unifiedfleet/api/v1/proto"
)

type DecoratedFleet struct {
	// Service is the service to decorate.
	Service FleetServer
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

func (s *DecoratedFleet) CreateChromePlatform(ctx context.Context, req *CreateChromePlatformRequest) (rsp *proto1.ChromePlatform, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "CreateChromePlatform", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.CreateChromePlatform(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "CreateChromePlatform", rsp, err)
	}
	return
}

func (s *DecoratedFleet) UpdateChromePlatform(ctx context.Context, req *UpdateChromePlatformRequest) (rsp *proto1.ChromePlatform, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "UpdateChromePlatform", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.UpdateChromePlatform(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "UpdateChromePlatform", rsp, err)
	}
	return
}

func (s *DecoratedFleet) GetChromePlatform(ctx context.Context, req *GetChromePlatformRequest) (rsp *proto1.ChromePlatform, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "GetChromePlatform", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.GetChromePlatform(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "GetChromePlatform", rsp, err)
	}
	return
}

func (s *DecoratedFleet) ListChromePlatforms(ctx context.Context, req *ListChromePlatformsRequest) (rsp *ListChromePlatformsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "ListChromePlatforms", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ListChromePlatforms(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "ListChromePlatforms", rsp, err)
	}
	return
}

func (s *DecoratedFleet) DeleteChromePlatform(ctx context.Context, req *DeleteChromePlatformRequest) (rsp *empty.Empty, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "DeleteChromePlatform", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.DeleteChromePlatform(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "DeleteChromePlatform", rsp, err)
	}
	return
}

func (s *DecoratedFleet) ImportChromePlatforms(ctx context.Context, req *ImportChromePlatformsRequest) (rsp *status.Status, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "ImportChromePlatforms", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ImportChromePlatforms(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "ImportChromePlatforms", rsp, err)
	}
	return
}

func (s *DecoratedFleet) CreateMachine(ctx context.Context, req *CreateMachineRequest) (rsp *proto1.Machine, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "CreateMachine", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.CreateMachine(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "CreateMachine", rsp, err)
	}
	return
}

func (s *DecoratedFleet) UpdateMachine(ctx context.Context, req *UpdateMachineRequest) (rsp *proto1.Machine, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "UpdateMachine", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.UpdateMachine(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "UpdateMachine", rsp, err)
	}
	return
}

func (s *DecoratedFleet) GetMachine(ctx context.Context, req *GetMachineRequest) (rsp *proto1.Machine, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "GetMachine", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.GetMachine(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "GetMachine", rsp, err)
	}
	return
}

func (s *DecoratedFleet) ListMachines(ctx context.Context, req *ListMachinesRequest) (rsp *ListMachinesResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "ListMachines", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ListMachines(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "ListMachines", rsp, err)
	}
	return
}

func (s *DecoratedFleet) DeleteMachine(ctx context.Context, req *DeleteMachineRequest) (rsp *empty.Empty, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "DeleteMachine", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.DeleteMachine(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "DeleteMachine", rsp, err)
	}
	return
}

func (s *DecoratedFleet) ImportMachines(ctx context.Context, req *ImportMachinesRequest) (rsp *status.Status, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "ImportMachines", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ImportMachines(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "ImportMachines", rsp, err)
	}
	return
}

func (s *DecoratedFleet) CreateRack(ctx context.Context, req *CreateRackRequest) (rsp *proto1.Rack, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "CreateRack", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.CreateRack(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "CreateRack", rsp, err)
	}
	return
}

func (s *DecoratedFleet) UpdateRack(ctx context.Context, req *UpdateRackRequest) (rsp *proto1.Rack, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "UpdateRack", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.UpdateRack(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "UpdateRack", rsp, err)
	}
	return
}

func (s *DecoratedFleet) GetRack(ctx context.Context, req *GetRackRequest) (rsp *proto1.Rack, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "GetRack", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.GetRack(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "GetRack", rsp, err)
	}
	return
}

func (s *DecoratedFleet) ListRacks(ctx context.Context, req *ListRacksRequest) (rsp *ListRacksResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "ListRacks", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ListRacks(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "ListRacks", rsp, err)
	}
	return
}

func (s *DecoratedFleet) DeleteRack(ctx context.Context, req *DeleteRackRequest) (rsp *empty.Empty, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "DeleteRack", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.DeleteRack(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "DeleteRack", rsp, err)
	}
	return
}

func (s *DecoratedFleet) CreateMachineLSE(ctx context.Context, req *CreateMachineLSERequest) (rsp *proto1.MachineLSE, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "CreateMachineLSE", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.CreateMachineLSE(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "CreateMachineLSE", rsp, err)
	}
	return
}

func (s *DecoratedFleet) UpdateMachineLSE(ctx context.Context, req *UpdateMachineLSERequest) (rsp *proto1.MachineLSE, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "UpdateMachineLSE", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.UpdateMachineLSE(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "UpdateMachineLSE", rsp, err)
	}
	return
}

func (s *DecoratedFleet) GetMachineLSE(ctx context.Context, req *GetMachineLSERequest) (rsp *proto1.MachineLSE, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "GetMachineLSE", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.GetMachineLSE(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "GetMachineLSE", rsp, err)
	}
	return
}

func (s *DecoratedFleet) ListMachineLSEs(ctx context.Context, req *ListMachineLSEsRequest) (rsp *ListMachineLSEsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "ListMachineLSEs", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ListMachineLSEs(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "ListMachineLSEs", rsp, err)
	}
	return
}

func (s *DecoratedFleet) DeleteMachineLSE(ctx context.Context, req *DeleteMachineLSERequest) (rsp *empty.Empty, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "DeleteMachineLSE", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.DeleteMachineLSE(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "DeleteMachineLSE", rsp, err)
	}
	return
}

func (s *DecoratedFleet) CreateRackLSE(ctx context.Context, req *CreateRackLSERequest) (rsp *proto1.RackLSE, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "CreateRackLSE", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.CreateRackLSE(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "CreateRackLSE", rsp, err)
	}
	return
}

func (s *DecoratedFleet) UpdateRackLSE(ctx context.Context, req *UpdateRackLSERequest) (rsp *proto1.RackLSE, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "UpdateRackLSE", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.UpdateRackLSE(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "UpdateRackLSE", rsp, err)
	}
	return
}

func (s *DecoratedFleet) GetRackLSE(ctx context.Context, req *GetRackLSERequest) (rsp *proto1.RackLSE, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "GetRackLSE", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.GetRackLSE(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "GetRackLSE", rsp, err)
	}
	return
}

func (s *DecoratedFleet) ListRackLSEs(ctx context.Context, req *ListRackLSEsRequest) (rsp *ListRackLSEsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "ListRackLSEs", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ListRackLSEs(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "ListRackLSEs", rsp, err)
	}
	return
}

func (s *DecoratedFleet) DeleteRackLSE(ctx context.Context, req *DeleteRackLSERequest) (rsp *empty.Empty, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "DeleteRackLSE", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.DeleteRackLSE(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "DeleteRackLSE", rsp, err)
	}
	return
}

func (s *DecoratedFleet) CreateNic(ctx context.Context, req *CreateNicRequest) (rsp *proto1.Nic, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "CreateNic", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.CreateNic(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "CreateNic", rsp, err)
	}
	return
}

func (s *DecoratedFleet) UpdateNic(ctx context.Context, req *UpdateNicRequest) (rsp *proto1.Nic, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "UpdateNic", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.UpdateNic(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "UpdateNic", rsp, err)
	}
	return
}

func (s *DecoratedFleet) GetNic(ctx context.Context, req *GetNicRequest) (rsp *proto1.Nic, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "GetNic", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.GetNic(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "GetNic", rsp, err)
	}
	return
}

func (s *DecoratedFleet) ListNics(ctx context.Context, req *ListNicsRequest) (rsp *ListNicsResponse, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "ListNics", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ListNics(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "ListNics", rsp, err)
	}
	return
}

func (s *DecoratedFleet) DeleteNic(ctx context.Context, req *DeleteNicRequest) (rsp *empty.Empty, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "DeleteNic", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.DeleteNic(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "DeleteNic", rsp, err)
	}
	return
}

func (s *DecoratedFleet) ImportNics(ctx context.Context, req *ImportNicsRequest) (rsp *status.Status, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "ImportNics", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ImportNics(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "ImportNics", rsp, err)
	}
	return
}

func (s *DecoratedFleet) ImportDatacenters(ctx context.Context, req *ImportDatacentersRequest) (rsp *status.Status, err error) {
	if s.Prelude != nil {
		var newCtx context.Context
		newCtx, err = s.Prelude(ctx, "ImportDatacenters", req)
		if err == nil {
			ctx = newCtx
		}
	}
	if err == nil {
		rsp, err = s.Service.ImportDatacenters(ctx, req)
	}
	if s.Postlude != nil {
		err = s.Postlude(ctx, "ImportDatacenters", rsp, err)
	}
	return
}
