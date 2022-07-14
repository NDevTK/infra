package templates

import (
	"errors"
	"testing"

	"go.chromium.org/chromiumos/config/go/test/lab/api"
)

type mockLookuper struct {
	ContainerLookuper
	portLookupFunc func(string) ([]*PortBinding, error)
	ipLookupFunc   func(string) (string, error)
}

func (m *mockLookuper) LookupContainerPortBindings(name string) ([]*PortBinding, error) {
	return m.portLookupFunc(name)
}

func (m *mockLookuper) LookupContainerIpAddress(name string) (string, error) {
	return m.ipLookupFunc(name)
}

func TestExtract_invalid(t *testing.T) {
	router := populatorRouter{containerLookuper: &mockLookuper{}}
	endpoint := api.IpEndpoint{Address: "test"}
	scheme, _, err := router.extract(endpoint)
	if scheme != "" {
		t.Fatalf("scheme should be empty for invalid endpoint")
	}
	if err == nil {
		t.Fatalf("error should be thrown")
	}
}

func TestExtract_hostPort(t *testing.T) {
	router := populatorRouter{containerLookuper: &mockLookuper{}}
	endpoint := api.IpEndpoint{Address: "ctr-host-port://container-name", Port: 1}
	expectedEndpoint := api.IpEndpoint{Address: "container-name", Port: 1}
	scheme, returnedEndpoint, err := router.extract(endpoint)
	if scheme != "ctr-host-port" {
		t.Fatalf("scheme does not match")
	}
	if err != nil {
		t.Fatalf("unexpectedError")
	}
	checkEndpoint(t, expectedEndpoint, returnedEndpoint)
}

func TestExtract_containerIp(t *testing.T) {
	router := populatorRouter{containerLookuper: &mockLookuper{}}
	endpoint := api.IpEndpoint{Address: "ctr-container-ip://container-name", Port: 1}
	expectedEndpoint := api.IpEndpoint{Address: "container-name", Port: 1}
	scheme, returnedEndpoint, err := router.extract(endpoint)
	if scheme != "ctr-container-ip" {
		t.Fatalf("scheme does not match")
	}
	if err != nil {
		t.Fatalf("unexpectedError")
	}
	checkEndpoint(t, expectedEndpoint, returnedEndpoint)
}

func TestPopulate_hostPort(t *testing.T) {
	expectedAddress := "localhost"
	expectedPort := 42222
	expectedContainerName := "container-name"
	expectedEndpoint := api.IpEndpoint{
		Address: expectedAddress,
		Port:    int32(expectedPort),
	}
	router := populatorRouter{containerLookuper: &mockLookuper{
		portLookupFunc: func(s string) ([]*PortBinding, error) {
			if s != expectedContainerName {
				t.Fatalf("container name does not match\nexpect: %s\nactual: %s",
					expectedContainerName, s)
			}
			return []*PortBinding{{HostPort: expectedPort}}, nil
		}}}
	endpoint := api.IpEndpoint{Address: "ctr-host-port://container-name", Port: 1}

	returnedEndpoint, err := router.populate(endpoint)

	if err != nil {
		t.Fatalf("unexpectedError")
	}
	checkEndpoint(t, expectedEndpoint, returnedEndpoint)
}

func TestPopulate_hostPort_error(t *testing.T) {
	expectedEndpoint := api.IpEndpoint{Address: "container-name", Port: 1}
	router := populatorRouter{containerLookuper: &mockLookuper{
		portLookupFunc: func(s string) ([]*PortBinding, error) {
			return nil, errors.New("command throw error")
		}}}
	endpoint := api.IpEndpoint{Address: "ctr-host-port://container-name", Port: 1}

	returnedEndpoint, err := router.populate(endpoint)

	if err == nil {
		t.Fatalf("expect error to be returned")
	}
	checkEndpoint(t, expectedEndpoint, returnedEndpoint)
}

func TestPopulate_hostPort_multiplePorts(t *testing.T) {
	expectedEndpoint := api.IpEndpoint{Address: "container-name", Port: 1}
	router := populatorRouter{containerLookuper: &mockLookuper{
		portLookupFunc: func(s string) ([]*PortBinding, error) {
			return []*PortBinding{{HostPort: 42}, {HostPort: 43}}, nil
		}}}
	endpoint := api.IpEndpoint{Address: "ctr-host-port://container-name", Port: 1}

	returnedEndpoint, err := router.populate(endpoint)

	if err == nil {
		t.Fatalf("expect error to be returned")
	}
	checkEndpoint(t, expectedEndpoint, returnedEndpoint)
}

func TestPopulate_containerIp(t *testing.T) {
	expectedAddress := "192.168.10.2"
	expectedPort := 1
	expectedContainerName := "container-name"
	expectedEndpoint := api.IpEndpoint{
		Address: expectedAddress,
		Port:    int32(expectedPort),
	}
	router := populatorRouter{containerLookuper: &mockLookuper{
		ipLookupFunc: func(s string) (string, error) {
			if s != expectedContainerName {
				t.Fatalf("container name does not match\nexpect: %s\nactual: %s",
					expectedContainerName, s)
			}
			return expectedAddress, nil
		}}}
	endpoint := api.IpEndpoint{
		Address: "ctr-container-ip://container-name",
		Port:    1,
	}

	returnedEndpoint, err := router.populate(endpoint)

	if err != nil {
		t.Fatalf("unexpectedError")
	}
	checkEndpoint(t, expectedEndpoint, returnedEndpoint)
}

func TestPopulate_containerIp_error(t *testing.T) {
	expectedEndpoint := api.IpEndpoint{Address: "container-name", Port: 1}
	router := populatorRouter{containerLookuper: &mockLookuper{
		ipLookupFunc: func(s string) (string, error) {
			return "", errors.New("command throw error")
		}}}
	endpoint := api.IpEndpoint{
		Address: "ctr-container-ip://container-name",
		Port:    1,
	}

	returnedEndpoint, err := router.populate(endpoint)

	if err == nil {
		t.Fatalf("expect error to be returned")
	}
	checkEndpoint(t, expectedEndpoint, returnedEndpoint)
}

func TestPopulate_containerIp_empty(t *testing.T) {
	expectedEndpoint := api.IpEndpoint{Address: "container-name", Port: 1}
	router := populatorRouter{containerLookuper: &mockLookuper{
		ipLookupFunc: func(s string) (string, error) {
			return "", nil
		}}}
	endpoint := api.IpEndpoint{
		Address: "ctr-container-ip://container-name",
		Port:    1,
	}

	returnedEndpoint, err := router.populate(endpoint)

	if err == nil {
		t.Fatalf("expect error to be returned")
	}
	checkEndpoint(t, expectedEndpoint, returnedEndpoint)
}

func checkEndpoint(t *testing.T, actual api.IpEndpoint, expect api.IpEndpoint) {
	if actual.Address != expect.Address || actual.Port != expect.Port {
		t.Fatalf("returned endpoint doesn't match\nexpect: %v\nactual: %v",
			expect, actual)
	}
}
