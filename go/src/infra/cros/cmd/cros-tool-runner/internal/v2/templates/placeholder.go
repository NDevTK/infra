package templates

import (
	"log"
	"regexp"

	"go.chromium.org/chromiumos/config/go/test/lab/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// placeholderPopulator is the interface to populate a placeholder IpEndpoint
// with actual values. The input uses URI schemes in address to indicate how the
// value should be populated. For example,
// ctr-host-port://container-name indicates address should be localhost and port
// should be the port mapped on the host
// ctr-container-ip://container-name indicates address should be the IP address
// of the container and port should be intact.
// A placeholderPopulator then populates the value using proper commands.
type placeholderPopulator interface {
	// populate takes IpEndpoint template and returns a new IpEndpoint with actual
	// values populated
	// InvalidArgument error should be returned if unable to populate.
	populate(api.IpEndpoint) (api.IpEndpoint, error)
}

// Scheme definitions
const (
	HostPortScheme    = "ctr-host-port"
	ContainerIpScheme = "ctr-container-ip"
)

// populatorRouter is the entry point
type populatorRouter struct {
	placeholderPopulator
	containerLookuper ContainerLookuper
}

func newPopulatorRouter() *populatorRouter {
	return &populatorRouter{containerLookuper: &TemplateUtils}
}

// extract returns scheme, a copy of IpEndpoint with address replaced by
// container name
func (pr *populatorRouter) extract(endpoint api.IpEndpoint) (string, api.IpEndpoint, error) {
	r := regexp.MustCompile(`(?P<Scheme>ctr-[\w-]+)://(?P<ContainerName>.+)`)
	match := r.FindStringSubmatch(endpoint.Address)
	if len(match) != 3 {
		return "", endpoint, status.Error(codes.InvalidArgument, "Not a valid template placeholder")
	}
	scheme := match[1]
	containerName := match[2]
	return scheme, api.IpEndpoint{Address: containerName, Port: endpoint.Port}, nil
}

func (pr *populatorRouter) populate(input api.IpEndpoint) (api.IpEndpoint, error) {
	scheme, updatedEndpoint, err := pr.extract(input)
	if err != nil {
		log.Printf("skip unpopulatable input %v", input)
		return input, err
	}
	switch scheme {
	case HostPortScheme:
		actualPopulator := hostPortPopulator{pr.containerLookuper}
		return actualPopulator.populate(updatedEndpoint)
	case ContainerIpScheme:
		actualPopulator := containerIpPopulator{pr.containerLookuper}
		return actualPopulator.populate(updatedEndpoint)
	default:
		return input, status.Error(codes.InvalidArgument, "Scheme is unrecognized")
	}
}

// hostPortPopulator populates host port using the updated IpEndpoint template
// supplied by the router
type hostPortPopulator struct {
	containerLookup ContainerLookuper
}

func (p *hostPortPopulator) populate(input api.IpEndpoint) (api.IpEndpoint, error) {
	ports, err := p.containerLookup.LookupContainerPortBindings(input.Address)
	if err != nil {
		return input, err
	}
	if len(ports) != 1 {
		return input, status.Error(codes.InvalidArgument, "The container has more than one port bindings")
	}
	return api.IpEndpoint{Address: "localhost", Port: int32(ports[0].HostPort)}, nil
}

// hostPortPopulator populates container ip using the updated IpEndpoint
// template supplied by the router
type containerIpPopulator struct {
	containerLookup ContainerLookuper
}

func (p *containerIpPopulator) populate(input api.IpEndpoint) (api.IpEndpoint, error) {
	ip, err := p.containerLookup.LookupContainerIpAddress(input.Address)
	if err != nil {
		return input, err
	}
	if ip == "" {
		return input, status.Error(codes.InvalidArgument, "The container does not have a valid IP returned")
	}
	return api.IpEndpoint{Address: ip, Port: input.Port}, nil
}
