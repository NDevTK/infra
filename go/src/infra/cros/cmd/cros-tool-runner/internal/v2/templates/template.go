package templates

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"infra/cros/cmd/cros-tool-runner/api"
)

// TemplateProcessor converts a container-specific template into a valid generic
// StartContainerRequest. Besides request conversions, a TemplateProcessor is
// also aware of a container's dependencies of other containers, whose addresses
// are determined at runtime. The addresses are provided as IpEndpoint
// placeholders in a template, and TemplateProcessor use placeholderPopulators
// to populate actual values.
type TemplateProcessor interface {
	Process(*api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error)
}

type RequestRouter struct{}

func (*RequestRouter) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	switch t := request.Template.Container.(type) {
	case *api.Template_CrosDut:
		actualProcessor := CrosDutProcessor{}
		return actualProcessor.Process(request)
	case *api.Template_CrosProvision:
		actualProcessor := CrosProvisionProcessor{}
		return actualProcessor.Process(request)
	case *api.Template_CrosTest:
		actualProcessor := CrosTestProcessor{}
		return actualProcessor.Process(request)
	default:
		return nil, status.Error(codes.Unimplemented, fmt.Sprintf("%v to be implemented", t))
	}
}

type CrosDutProcessor struct{}

func (*CrosDutProcessor) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	// TODO(b/237695921) implement template processor for cros-dut, create new file if necessary
	if request.Template.GetCrosDut() == nil {
		return nil, status.Error(codes.Internal, "unable to process")
	}
	return &api.StartContainerRequest{Name: request.Name, ContainerImage: request.ContainerImage}, nil
}

type CrosProvisionProcessor struct{}

func (*CrosProvisionProcessor) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	// TODO(b/237696880) implement template processor for cros-provision, create new file if necessary
	if request.Template.GetCrosProvision() == nil {
		return nil, status.Error(codes.Internal, "unable to process")
	}
	return &api.StartContainerRequest{Name: request.Name, ContainerImage: request.ContainerImage}, nil
}

type CrosTestProcessor struct{}

func (*CrosTestProcessor) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	// TODO(b/237696881) implement template processor for cros-test, create new file if necessary
	if request.Template.GetCrosTest() == nil {
		return nil, status.Error(codes.Internal, "unable to process")
	}
	return &api.StartContainerRequest{Name: request.Name, ContainerImage: request.ContainerImage}, nil
}
