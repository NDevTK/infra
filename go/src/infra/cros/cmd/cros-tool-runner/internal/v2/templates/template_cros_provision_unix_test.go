//go:build !windows
// +build !windows

package templates

import (
	"testing"

	_go "go.chromium.org/chromiumos/config/go"
	testApi "go.chromium.org/chromiumos/config/go/test/api"
	labApi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"infra/cros/cmd/cros-tool-runner/api"
)

func TestCrosProvisionPopulate(t *testing.T) {
	processor := newCrosProvisionProcessor()
	expect := labApi.IpEndpoint{Address: "localhost", Port: 80}
	processor.placeholderPopulator = newMockWithEndpoint(&expect)
	request := &api.StartTemplatedContainerRequest{
		Name:           "my-container",
		ContainerImage: "gcr.io/image:123",
		Template: &api.Template{
			Container: &api.Template_CrosProvision{
				CrosProvision: &api.CrosProvisionTemplate{
					Network:     "mynet",
					ArtifactDir: "/tmp",
					InputRequest: &testApi.CrosProvisionRequest{
						DutServer: &labApi.IpEndpoint{Address: "ctr-host-port://dut-name", Port: 0},
						Dut:       &labApi.Dut{Id: &labApi.Dut_Id{Value: "chromeos6-row4-rack5-host14"}},
						ProvisionState: &testApi.ProvisionState{SystemImage: &testApi.ProvisionState_SystemImage{
							SystemImagePath: &_go.StoragePath{
								Path:     "gs://chromeos-image-archive/kevin-cq/R104-14895.0.0-66173-8812350496939596961",
								HostType: _go.StoragePath_GS,
							}}},
					}}}}}

	convertedRequest, err := processor.Process(request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	check(t, convertedRequest.Name, request.Name)
	check(t, convertedRequest.ContainerImage, request.ContainerImage)
	check(t, convertedRequest.AdditionalOptions.Network, "mynet")
	check(t, convertedRequest.AdditionalOptions.Expose[0], "80")
	check(t, convertedRequest.AdditionalOptions.Volume[0], "/tmp:/tmp/provisionservice")
	check(t, convertedRequest.StartCommand[0], "cros-provision")
}
