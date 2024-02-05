// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/cros-tool-runner/internal/v2/commands"

	"go.chromium.org/chromiumos/config/go/test/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type crosTestProcessor struct {
	cmdExecutor       cmdExecutor
	defaultServerPort string // Default port used in cros-test
}

func newCrosTestProcessor() *crosTestProcessor {
	return &crosTestProcessor{
		cmdExecutor:       &commands.ContextualExecutor{},
		defaultServerPort: "8001",
	}
}

func (p *crosTestProcessor) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	t := request.GetTemplate().GetCrosTest()
	if t == nil {
		return nil, status.Error(codes.Internal, "unable to process")
	}

	port := portZero
	expose := make([]string, 0)
	if request.Network != hostNetworkName {
		port = p.defaultServerPort
		expose = append(expose, port)
	}
	// All non-test harness artifacts will be in <artifact_dir>/cros-test/cros-test.
	crosTestDir := path.Join(request.ArtifactDir, "cros-test", "cros-test")
	// All test result artifacts will be in <artifact_dir>/cros-test/results.
	resultDir := path.Join(request.ArtifactDir, "cros-test", "results")
	// Setting up directories. Required as podman doesn't create directories for volume mounting.
	p.createDir(crosTestDir)
	p.createDir(resultDir)
	volumes := []string{
		fmt.Sprintf("%s:%s", crosTestDir, "/tmp/test/cros-test"),
		fmt.Sprintf("%s:%s", resultDir, "/tmp/test/results"),
	}
	// Mount authorization file for gsutil if exists. See b/239855913
	gsutilAuthFile := "/home/chromeos-test/.boto"
	if _, err := os.Stat(gsutilAuthFile); err == nil {
		volumes = append(volumes, fmt.Sprintf("%s:%s", gsutilAuthFile, gsutilAuthFile))
	}
	// Mount autotest results shared folder if exists. See b/239855163
	autotestResultsFolder := "/usr/local/autotest/results/shared"
	if _, err := os.Stat(autotestResultsFolder); err == nil {
		volumes = append(volumes, fmt.Sprintf("%s:%s", autotestResultsFolder, autotestResultsFolder))
	}
	// Mount docker socket file if it exists this allows tests to orchestrate servod containers if needed
	if botProvider := common.GetBotProvider(); botProvider == common.BotProviderPVS {
		const dockerSock = "/var/run/docker.sock"
		if _, err := os.Stat(dockerSock); err == nil {
			volumes = append(volumes, fmt.Sprintf("%s:%s", dockerSock, dockerSock))
		}
	}
	if _, err := os.Stat(HostServiceAcctCredsDir); err == nil {
		volumes = append(volumes, fmt.Sprintf("%s:%s", HostServiceAcctCredsDir, HostServiceAcctCredsDir))
	}

	additionalOptions := &api.StartContainerRequest_Options{
		Network: request.Network,
		Expose:  expose,
		Volume:  volumes,
	}
	// It is necessary to do sudo here because /tmp/test is owned by root inside docker
	// when docker mount /tmp/test. However, the user that is running cros-test is
	// chromeos-test inside docker. Hence, the user chromeos-test does not have write
	// permission in /tmp/test. Therefore, we need to change the owner of the directory.
	cmd := fmt.Sprintf("sudo --non-interactive chown -R chromeos-test:chromeos-test %s && cros-test server -port %s", "/tmp/test", port)
	startCommand := []string{"bash", "-c", cmd}
	if botProvider := common.GetBotProvider(); botProvider == common.BotProviderPVS {
		if localDockerGID, err := user.LookupGroup("docker"); err == nil {
			log.Printf("gid found for docker %+v, updating GID in container", localDockerGID)
			// modify the docker group GID to match inside the container
			// this allows chromeos-test to have permissions on dockerSock
			// we must start a new shell with su after changing group ID for it to be recognized
			altCmd := fmt.Sprintf("sudo --non-interactive groupadd -g %s docker && "+
				"sudo --non-interactive usermod -a -G docker chromeos-test && "+
				"sudo --non-interactive su - chromeos-test -c '%s'",
				localDockerGID.Gid,
				cmd,
			)
			startCommand = []string{"bash", "-c", altCmd}
		} else {
			log.Printf("no gid found for docker: %+v, skipping docker GID update", err)
		}
	}
	return &api.StartContainerRequest{Name: request.Name, ContainerImage: request.ContainerImage, AdditionalOptions: additionalOptions, StartCommand: startCommand}, nil
}

func (p *crosTestProcessor) discoverPort(request *api.StartTemplatedContainerRequest) (*api.Container_PortBinding, error) {
	// delegate to default impl, any template-specific logic should be implemented here.
	return defaultDiscoverPort(p.cmdExecutor, request)
}

// createDir creates artifact subdirectories for the given path.
func (p *crosTestProcessor) createDir(dirPath string) {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Printf("warning: cros-test template processor received error when creating directory %s: %v", dirPath, err)
	}
	log.Printf("cros-test template processor has created directory %s", dirPath)
}
