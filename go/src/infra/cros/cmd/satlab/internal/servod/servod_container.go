// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servod

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/docker"
)

// DockerClient is an interface fulfilled by the recovery docker lib
// Used here to facilitate testing
type DockerClient interface {
	Start(ctx context.Context, containerName string, req *docker.ContainerArgs, timeout time.Duration) (*docker.StartResponse, error)
	IsUp(ctx context.Context, containerName string) (bool, error)
}

// StartServodContainer is used to start the docker container for servod
// If there is already a container running with the same name it will not start a new container
func StartServodContainer(d DockerClient, ctx context.Context, servoContainerName string, board string, model string, servoSerial string) (*docker.StartResponse, error) {
	// check presence of running container already
	if up, err := d.IsUp(ctx, servoContainerName); err != nil {
		return nil, err
	} else if up {
		return nil, errors.Reason("Docker container with name %s is already running", servoContainerName).Err()
	}

	args := buildServodDockerArgs(servoContainerName, board, model, servoSerial)

	res, err := d.Start(ctx, servoContainerName, args, time.Minute)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Launched container. To access, run:\n\tdocker exec -it %s bash\n", strings.TrimSuffix(res.Stdout, "\n"))
	return res, nil
}

// buildServodDockerArgs produces ContainerArgs which has the full information needed to spin up a servod container via `docker run ...`
func buildServodDockerArgs(servoContainerName string, board string, model string, servoSerial string) *docker.ContainerArgs {
	return &docker.ContainerArgs{
		Detached:   true,
		ImageName:  dockerServodImageName(),
		EnvVar:     generateEnvVars(board, model, servoSerial),
		Volumes:    generateVols(servoSerial),
		Network:    "default_satlab",
		Privileged: true,
		Exec:       []string{"bash", "/start_servod.sh"},
	}
}

// generateEnvVars builds a string array of env vars needed to launch servod in docker
func generateEnvVars(board string, model string, servoSerial string) []string {
	port := 9999
	var envVars []string

	envVars = append(envVars, fmt.Sprintf("BOARD=%s", board))
	envVars = append(envVars, fmt.Sprintf("MODEL=%s", model))
	envVars = append(envVars, fmt.Sprintf("SERVO_SERIAL=%s", servoSerial))
	envVars = append(envVars, fmt.Sprintf("PORT=%d", port))

	return envVars
}

// generateVols builds a string array of volumes needed to launch servod in docker
func generateVols(servoContainerName string) []string {
	var vols []string

	vols = append(vols, "/dev:/dev")
	vols = append(vols, fmt.Sprintf("%s_log:/var/log/servod_9999/", servoContainerName))

	return vols
}

// dockerServodImageName builds the appropriate image name for servod based on env vars
// duplicates logic in TLW client
func dockerServodImageName() string {
	// TODO(elijahtrexler) add these variables to SATLAB_REMOTE_ACCESS
	label := getEnv("SERVOD_CONTAINER_LABEL", "release")
	registry := getEnv("REGISTRY_URI", "us-docker.pkg.dev/chromeos-partner-moblab/common-core")
	return fmt.Sprintf("%s/servod:%s", registry, label)
}

// getEnv is helper to get env variables and falling back if not set
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
