package templates

import (
	"strings"

	"go.chromium.org/chromiumos/config/go/test/lab/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// placeholderPopulator is the interface to populate a placeholder IpEndpoint
// with actual values. The input has a prefix in address to indicate how the
// value should be populated. For example,
// host-port://container-name indicates address should be localhost and port
// should be the port mapped on the host
// container-ip://container-name indicates address should be the IP address of
// the container and port should be intact.
// A placeholderPopulator then populates the value using proper commands.
type placeholderPopulator interface {
	// populate takes IpEndpoint template and returns IpEndpoint with actual value
	// InvalidArgument error should be returned if unable to populate.
	populate(*api.IpEndpoint) (*api.IpEndpoint, error)
}

// extract returns prefix, container name, and port number from an IpEndpoint template
func extract(endpoint *api.IpEndpoint) (string, string, int32) {
	// TODO(mingkong) implement extract
	return "", endpoint.Address, endpoint.Port
}

type populatorRouter struct{}

func (*populatorRouter) populate(input *api.IpEndpoint) (*api.IpEndpoint, error) {
	// TODO(mingkong): finalize prefixes and define constants.
	switch {
	case strings.HasPrefix(input.Address, "host-port://"):
		actualPopulator := hostPortPopulator{}
		return actualPopulator.populate(input)
	case strings.HasPrefix(input.Address, "container-ip://"):
		actualPopulator := containerIpPopulator{}
		return actualPopulator.populate(input)
	default:
		return input, status.Error(codes.InvalidArgument, "Not a valid template placeholder")
	}
}

type hostPortPopulator struct{}

func (*hostPortPopulator) populate(input *api.IpEndpoint) (*api.IpEndpoint, error) {
	// TODO(mingkong) implement hostPortPopulator, execute proper command in commands package to lookup
	// e.g. docker container inspect --format
	return input, nil
}

type containerIpPopulator struct{}

func (*containerIpPopulator) populate(input *api.IpEndpoint) (*api.IpEndpoint, error) {
	// TODO(mingkong) implement containerIpPopulator, execute proper command in commands package to lookup
	// e.g. docker container inspect --format
	return input, nil
}
