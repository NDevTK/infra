package templates

import (
	"errors"
	"testing"

	testApi "go.chromium.org/chromiumos/config/go/test/api"
	labApi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"infra/cros/cmd/cros-tool-runner/api"
)

type mockPlaceholderPopulator struct {
	placeholderPopulator
	populateFunc func(labApi.IpEndpoint) (labApi.IpEndpoint, error)
}

func (m *mockPlaceholderPopulator) populate(endpoint labApi.IpEndpoint) (labApi.IpEndpoint, error) {
	return m.populateFunc(endpoint)
}

func newMockWithError() *mockPlaceholderPopulator {
	return &mockPlaceholderPopulator{
		populateFunc: func(endpoint labApi.IpEndpoint) (labApi.IpEndpoint, error) {
			return endpoint, errors.New("some error")
		}}
}

func newMockWithEndpoint(expect *labApi.IpEndpoint) *mockPlaceholderPopulator {
	return &mockPlaceholderPopulator{
		populateFunc: func(endpoint labApi.IpEndpoint) (labApi.IpEndpoint, error) {
			return *expect, nil
		}}
}

func TestProcessPlaceholders(t *testing.T) {
	processor := newCrosProvisionProcessor()
	expect := labApi.IpEndpoint{Address: "localhost", Port: 12345}
	processor.placeholderPopulator = newMockWithEndpoint(&expect)
	request := &api.StartTemplatedContainerRequest{
		Template: &api.Template{
			Container: &api.Template_CrosProvision{
				CrosProvision: &api.CrosProvisionTemplate{
					InputRequest: &testApi.CrosProvisionRequest{
						DutServer: &labApi.IpEndpoint{Address: "ctr-host-port://dut-name", Port: 0},
					}}}}}

	processor.processPlaceholders(request)

	actual := request.Template.GetCrosProvision().InputRequest.DutServer
	if actual.Address != expect.Address || actual.Port != expect.Port {
		t.Fatalf("IpEndpoint wasn't populated  %s.", actual)
	}
}

func TestProcessPlaceholders_errorIgnored(t *testing.T) {
	processor := newCrosProvisionProcessor()
	expect := labApi.IpEndpoint{Address: "dut-name", Port: 0}
	processor.placeholderPopulator = newMockWithError()
	request := &api.StartTemplatedContainerRequest{
		Template: &api.Template{
			Container: &api.Template_CrosProvision{
				CrosProvision: &api.CrosProvisionTemplate{
					InputRequest: &testApi.CrosProvisionRequest{
						DutServer: &labApi.IpEndpoint{Address: "dut-name", Port: 0},
					}}}}}
	processor.processPlaceholders(request)

	actual := request.Template.GetCrosProvision().InputRequest.DutServer
	if actual.Address != expect.Address || actual.Port != expect.Port {
		t.Fatalf("IpEndpoint wasn't populated  %s.", actual)
	}
}

func check(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%v should match %v", a, b)
	}
}
