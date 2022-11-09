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

func getPortBindingErrorMessage(numberOfPortBindings int) string {
	if numberOfPortBindings == 0 {
		return "The container doesn't have any port bindings. Make sure to expose the service port when start container."
	}
	if numberOfPortBindings > 1 {
		return "The container has more than one port bindings. Make sure to expose only one service port when start container."
	}
	return ""
}
