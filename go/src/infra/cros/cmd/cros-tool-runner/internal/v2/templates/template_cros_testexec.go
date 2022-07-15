package templates

import (
	"fmt"
	"path"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"infra/cros/cmd/cros-tool-runner/api"
)

type crosTestProcessor struct{ TemplateProcessor }

func newCrosTestProcessor() TemplateProcessor {
	return &crosTestProcessor{}
}

func (p *crosTestProcessor) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	t := request.Template.GetCrosTest()
	if t == nil {
		return nil, status.Error(codes.Internal, "unable to process")
	}

	serverPort := "8001"
	// All non-test harness artifacts will be in <artifact_dir>/cros-test/cros-test.
	crosTestDir := path.Join(t.ArtifactDir, "cros-test", "cros-test")
	// All test result artifacts will be in <artifact_dir>/cros-test/results.
	resultDir := path.Join(t.ArtifactDir, "cros-test", "results")
	volumes := []string{
		fmt.Sprintf("%s:%s", crosTestDir, "/tmp/test/cros-test"),
		fmt.Sprintf("%s:%s", resultDir, "/tmp/test/results"),
	}
	additionalOptions := &api.StartContainerRequest_Options{
		Network: t.Network,
		Expose:  []string{serverPort},
		Volume:  volumes,
	}
	// It is necessary to do sudo here because /tmp/test is owned by root inside docker
	// when docker mount /tmp/test. However, the user that is running cros-test is
	// chromeos-test inside docker. Hence, the user chromeos-test does not have write
	// permission in /tmp/test. Therefore, we need to change the owner of the directory.
	cmd := fmt.Sprintf("sudo --non-interactive chown -R chromeos-test:chromeos-test %s && cros-test server", "/tmp/test")
	startCommand := []string{"bash", "-c", cmd}
	return &api.StartContainerRequest{Name: request.Name, ContainerImage: request.ContainerImage, AdditionalOptions: additionalOptions, StartCommand: startCommand}, nil
}
