package templates

import (
	"testing"

	labApi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"infra/cros/cmd/cros-tool-runner/api"
)

func TestCrosDutPopulate(t *testing.T) {
	processor := newCrosDutProcessor()
	request := &api.StartTemplatedContainerRequest{
		Name:           "my-container",
		ContainerImage: "gcr.io/image:123",
		Template: &api.Template{
			Container: &api.Template_CrosDut{
				CrosDut: &api.CrosDutTemplate{
					Network:     "mynet",
					ArtifactDir: "/tmp",
					CacheServer: &labApi.IpEndpoint{Address: "192.168.1.5", Port: 33},
					DutAddress:  &labApi.IpEndpoint{Address: "chromeos6-row4-rack5-host14", Port: 22},
				}}}}

	convertedRequest, err := processor.Process(request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	check(t, convertedRequest.Name, request.Name)
	check(t, convertedRequest.ContainerImage, request.ContainerImage)
	check(t, convertedRequest.AdditionalOptions.Network, "mynet")
	check(t, convertedRequest.AdditionalOptions.Expose[0], "80")
	check(t, convertedRequest.AdditionalOptions.Volume[0], "/tmp:/tmp/cros-dut")
	check(t, convertedRequest.StartCommand[0], "cros-dut")
}
