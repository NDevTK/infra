// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

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
// value should be populated.
// A placeholderPopulator then populates the value using proper commands.
type placeholderPopulator interface {
	// populate takes IpEndpoint template and returns a new IpEndpoint with actual
	// values populated
	// InvalidArgument error should be returned if unable to populate.
	populate(api.IpEndpoint) (api.IpEndpoint, error)
}

// Scheme definitions
const (
	// ContainerPortScheme indicates the port number used in the container (and
	// exposed and published when docker run) need to be populated into the
	// template. For example, if the service running in the container listens to
	// port 80 and the port has been exposed and published during docker run.
	// An IpEndpoint address of `ctr-container-port://container-name` indicates
	// that the address should be replaced with `container-name` and the port
	// should be replaced with actual port `80`. (Note that within a network, a
	// container can be referenced by name.)
	// To use this scheme, the port number in the input IpEndpoint must be 0.
	ContainerPortScheme = "ctr-container-port"
	// ContainerIpScheme is an experimental schema that indicates the container IP
	// need to be populated into the template. For example, if the container IP
	// address is 10.88.0.2. An IpEndpoint address of
	// `ctr-container-ip://container-name` will be replaced with `10.88.0.2`. If
	// port number is 0, it will be replaced with the container port similar to
	// the behavior of ctr-container-port.
	// Using container IP allows a service to be accessed outside its network.
	// Note that a container of podman must join a network to have an IP returned.
	ContainerIpScheme = "ctr-container-ip"
	// HostIpScheme is an experimental schema that indicates the host IP need to
	// be populated into the template. If port is 0 the template, it will
	// replaced with the host port found in the container's port binding.
	// Using host IP allows a service to be accessed outside its network.
	HostIpScheme = "ctr-host-ip"
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
	case ContainerPortScheme:
		actualPopulator := containerPortPopulator{pr.containerLookuper}
		return actualPopulator.populate(updatedEndpoint)
	case ContainerIpScheme:
		actualPopulator := containerIpPopulator{pr.containerLookuper}
		return actualPopulator.populate(updatedEndpoint)
	case HostIpScheme:
		actualPopulator := hostIpPopulator{pr.containerLookuper}
		return actualPopulator.populate(updatedEndpoint)
	default:
		return input, status.Error(codes.InvalidArgument, "Scheme is unrecognized")
	}
}

// containerPortPopulator populates container port using the updated IpEndpoint
// template supplied by the router
type containerPortPopulator struct {
	containerLookup ContainerLookuper
}

func (p *containerPortPopulator) populate(input api.IpEndpoint) (api.IpEndpoint, error) {
	ports, err := p.containerLookup.LookupContainerPortBindings(input.Address)
	if err != nil {
		return input, err
	}
	if len(ports) != 1 {
		return input, status.Error(codes.FailedPrecondition, getPortBindingErrorMessage(len(ports)))
	}
	if input.Port != 0 {
		return input, status.Error(codes.InvalidArgument, "The port number must be 0 to be used with ctr-container-port scheme")
	}
	return api.IpEndpoint{Address: input.Address, Port: int32(ports[0].ContainerPort)}, nil
}

// containerIpPopulator populates container IP (and port if 0) using the updated
// IpEndpoint template supplied by the router
type containerIpPopulator struct {
	containerLookup ContainerLookuper
}

func (p *containerIpPopulator) populate(input api.IpEndpoint) (api.IpEndpoint, error) {
	ip, err := p.containerLookup.LookupContainerIpAddress(input.Address)
	if err != nil {
		return input, err
	}
	if ip == "" {
		return input, status.Error(codes.FailedPrecondition, "The container does not have a valid IP returned. Make sure to specify network when start container.")
	}
	var port = input.Port
	if port == 0 {
		ports, err := p.containerLookup.LookupContainerPortBindings(input.Address)
		if err != nil {
			return input, err
		}
		if len(ports) != 1 {
			return input, status.Error(codes.FailedPrecondition, getPortBindingErrorMessage(len(ports)))
		}
		port = ports[0].ContainerPort
	}
	return api.IpEndpoint{Address: ip, Port: port}, nil
}

// hostIpPopulator populates host IP (and port if 0) using the updated
// IpEndpoint template supplied by the router
type hostIpPopulator struct {
	containerLookup ContainerLookuper
}

func (p *hostIpPopulator) populate(input api.IpEndpoint) (api.IpEndpoint, error) {
	ip, err := p.containerLookup.LookupHostIpAddress()
	if err != nil {
		return input, err
	}
	if ip == "" {
		return input, status.Error(codes.InvalidArgument, "Unable to retrieve the host IP")
	}
	var port = input.Port
	if port == 0 {
		ports, err := p.containerLookup.LookupContainerPortBindings(input.Address)
		if err != nil {
			return input, err
		}
		if len(ports) != 1 {
			return input, status.Error(codes.FailedPrecondition, getPortBindingErrorMessage(len(ports)))
		}
		port = ports[0].HostPort
	}
	return api.IpEndpoint{Address: ip, Port: port}, nil
}

func getPortBindingErrorMessage(numberOfPortBindings int) string {
	if numberOfPortBindings == 0 {
		return "The container doesn't have any port bindings. Make sure to expose the service port when start container."
	}
	if numberOfPortBindings > 1 {
		return "The container has more than one port bindings. Make sure to expose only one service port when start container."
	}
	return ""
}
