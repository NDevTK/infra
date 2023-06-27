// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executor

import (
	"context"
	"fmt"
	"infra/cros/cmd/cros_test_platformV2/containers"
	managers "infra/cros/cmd/cros_test_platformV2/docker_managers"
	"infra/cros/cmd/cros_test_runner/common"

	buildapi "go.chromium.org/chromiumos/config/go/build/api"

	"go.chromium.org/chromiumos/config/go/test/api"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/grpc"
)

type FilterExecutor struct {
	Ctr managers.ContainerManager

	conn *grpc.ClientConn

	binaryName    string
	containerPath string
}

func newFilterExecutor(ctr managers.ContainerManager, req *api.Filter, containerMetadata map[string]*buildapi.ContainerImageInfo) (*FilterExecutor, error) {
	var err error
	// For non-default filters, the given request might not include the path. If not try to find
	// it from the containermetadata.
	if req.Container.ContainerPath == "" {
		req, err = ResolvedContainer(req.Container.ServiceName, containerMetadata)
		if err != nil {
			return nil, err
		}
	}
	return &FilterExecutor{Ctr: ctr, binaryName: req.Container.ServiceName, containerPath: req.Container.ContainerPath}, nil
}

func (ex *FilterExecutor) Execute(ctx context.Context, cmd string, resp *TestPlanResponse) error {
	if cmd == "run" {
		return nil // Call the (running) binary inside the executing container.
	} else if cmd == "init" {
		fmt.Println("FILTER INIT!")
		ex.init()
		// TODO, consider moving this to "process", and adjusting process to meet our needs.
		return nil
	} else if cmd == "stop" {
		return nil // Stop containers.
	}
	return fmt.Errorf("invalid command given: %s\n", cmd)
}
func (ex *FilterExecutor) run() error {
	// Call binary, put the resp in ex.resp = ...
	return nil
}

// init starts the container, creates a client.
func (ex *FilterExecutor) init() error {
	fmt.Println("Starting container")
	ctx := context.Background()
	// Call build the genericTempaltedContainer interface.
	container := containers.NewContainer(ex.binaryName, ex.containerPath, ex.Ctr)
	template := &api.Template{Container: &api.Template_Generic{
		Generic: &testapi.GenericTemplate{
			DockerArtifactDir: fmt.Sprintf("/tmp/%s", ex.binaryName),
			BinaryArgs: []string{
				"server", "-port", "0",
			},
			BinaryName: ex.binaryName,
		},
	}}

	// Process does the init, run, getServer.
	serverAddress, err := container.ProcessContainer(ctx, template)
	if err != nil {
		fmt.Printf("error processing container:%s \n", err)

		return errors.Annotate(err, "error processing container: ").Err()
	}
	fmt.Println("Started container")

	// Connect with the service.
	conn, err := common.ConnectWithService(ctx, serverAddress)
	if err != nil {
		fmt.Printf(
			"error during connecting with %s server at %s: %s",
			ex.binaryName,
			serverAddress,
			err.Error())
		return err
	}
	ex.conn = conn

	// Current question: How do we know how to make a connection to the client?
	// Abstract client interface? Has to be...
	// Below is an example of making a conn to TestFinder...
	testClient := api.NewTestFinderServiceClient(conn)
	if testClient == nil {
		return fmt.Errorf("testFinderServiceClient is nil")
	}

	return nil

}
