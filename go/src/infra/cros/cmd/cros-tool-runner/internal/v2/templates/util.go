package templates

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"infra/cros/cmd/cros-tool-runner/internal/v2/commands"
)

// ContainerLookuper provides interface to lookup information for a container
type ContainerLookuper interface {
	LookupContainerPortBindings(name string) ([]*PortBinding, error)
	LookupContainerIpAddress(name string) (string, error)
}

// PortBinding represents a port mapping set by `docker run --publish`
// HostPort is the only information we are interested in.
type PortBinding struct {
	ContainerPort int
	Protocol      string
	HostIp        string
	HostPort      int
}

// templateUtils implements ContainerLookuper
type templateUtils struct {
	ContainerLookuper
}

var TemplateUtils = templateUtils{}

// parsePortBindingString parses the output from `docker container port` command
// The input string example: `81/tcp -> 0.0.0.0:42223`
func (*templateUtils) parsePortBindingString(input string) (*PortBinding, error) {
	r := regexp.MustCompile(`(?P<ContainerPort>\d+)/(?P<Protocol>\w+) -> (?P<HostIp>[\d\\.]+):(?P<HostPort>\d+)`)
	match := r.FindStringSubmatch(input)
	containerPort, err := strconv.Atoi(match[1])
	if err != nil {
		return nil, err
	}
	hostPort, err := strconv.Atoi(match[4])
	if err != nil {
		return nil, err
	}
	return &PortBinding{
		ContainerPort: containerPort,
		Protocol:      match[2],
		HostIp:        match[3],
		HostPort:      hostPort,
	}, nil
}

// parseMultilinePortBindings parses multiline output from `docker container
// port` command since Docker allows multiple ports to be published in one
// container. However, the CTRv2 server only allows one port to be published.
func (u *templateUtils) parseMultilinePortBindings(multiline string) ([]*PortBinding, error) {
	result := make([]*PortBinding, 0)
	for _, line := range strings.Split(multiline, "\n") {
		if line == "" {
			continue
		}
		binding, err := u.parsePortBindingString(line)
		if err != nil {
			return result, err
		}
		result = append(result, binding)
	}
	return result, nil
}

func (*templateUtils) retrieveContainerPortOutputFromCommand(name string) (string, error) {
	cmd := commands.ContainerPort{Name: name}
	stdout, _, err := cmd.Execute(context.Background())
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(stdout), err
}

// LookupContainerPortBindings is the API to get port bindings for a container
func (u *templateUtils) LookupContainerPortBindings(name string) ([]*PortBinding, error) {
	output, err := u.retrieveContainerPortOutputFromCommand(name)
	if err != nil {
		return nil, err
	}
	return u.parseMultilinePortBindings(output)
}

// LookupContainerIpAddress is the API to get the IP address of a container
func (*templateUtils) LookupContainerIpAddress(name string) (string, error) {
	cmd := commands.ContainerInspect{
		Names:  []string{name},
		Format: "'{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'",
	}
	stdout, _, err := cmd.Execute(context.Background())
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(stdout), err
}

// endpointToAddress converts an endpoint to an address string
func (*templateUtils) endpointToAddress(endpoint *api.IpEndpoint) string {
	return fmt.Sprintf("%s:%d", endpoint.Address, endpoint.Port)
}

// writeToFile writes proto message to a file
func (*templateUtils) writeToFile(file string, content proto.Message) error {
	f, err := os.Create(file)
	if err != nil {
		return errors.Annotate(err, "fail to create file %v", file).Err()
	}
	m := jsonpb.Marshaler{}
	if err := m.Marshal(f, content); err != nil {
		return errors.Annotate(err, "fail to marshal request to file %v", file).Err()
	}
	return nil
}
