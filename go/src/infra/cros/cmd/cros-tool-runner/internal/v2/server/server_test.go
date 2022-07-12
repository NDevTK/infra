package server

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"infra/cros/cmd/cros-tool-runner/api"
	"infra/cros/cmd/cros-tool-runner/internal/v2/commands"
)

func getService(executor CommandExecutor) *ContainerServerImpl {
	return &ContainerServerImpl{
		executor: executor,
	}
}

type mockExecutor struct {
	commandsExecuted     []string          // stores command type names. e.g. *commands.DockerRun
	commandsToThrowError map[string]string // tells executor to throw error. key is command type name, value is error message in stderr
}

func (m *mockExecutor) Execute(ctx context.Context, cmd commands.Command) (string, string, error) {
	cmdType := reflect.TypeOf(cmd).String()
	if errMsg, ok := m.commandsToThrowError[cmdType]; ok {
		return "", errMsg, errors.New(errMsg)
	}

	m.commandsExecuted = append(m.commandsExecuted, cmdType)
	return fmt.Sprintf("mockExecutor: %s executed", cmdType), "", nil
}

func TestCreateNetwork_invalidArgument_missingName(t *testing.T) {
	service := getService(&mockExecutor{})
	_, err := service.CreateNetwork(context.Background(), &api.CreateNetworkRequest{})
	if err == nil {
		t.Fatalf("Expect invalidArgument error")
	}
}

func TestCreateNetwork_cannotRetrieveId(t *testing.T) {
	errorMapping := make(map[string]string)
	errorMapping["*commands.NetworkList"] = "Some unknown error"
	executor := mockExecutor{commandsToThrowError: errorMapping}
	service := getService(&executor)
	_, err := service.CreateNetwork(context.Background(), &api.CreateNetworkRequest{Name: "mynet"})
	if err == nil {
		t.Fatalf("Expect notFound error")
	}
}

func TestCreateNetwork_success(t *testing.T) {
	executor := mockExecutor{}
	service := getService(&executor)
	_, err := service.CreateNetwork(context.Background(), &api.CreateNetworkRequest{Name: "mynet"})
	if err != nil {
		t.Fatalf("Expect success")
	}
	if len(executor.commandsExecuted) != 2 {
		t.Fatalf("Expect 2 commands have been executed")
	}
	if executor.commandsExecuted[0] != "*commands.NetworkCreate" {
		t.Fatalf("Expect network create have been executed")
	}
	if executor.commandsExecuted[1] != "*commands.NetworkList" {
		t.Fatalf("Expect network list have been executed")
	}
}

func TestStartContainer_invalidArgument_missingName(t *testing.T) {
	service := getService(&mockExecutor{})
	_, err := service.StartContainer(context.Background(), &api.StartContainerRequest{})
	if err == nil {
		t.Fatalf("Expect invalidArgument error")
	}
}

func TestStartContainer_invalidArgument_missingContainerImage(t *testing.T) {
	service := getService(&mockExecutor{})
	_, err := service.StartContainer(context.Background(), &api.StartContainerRequest{Name: "my-container"})
	if err == nil {
		t.Fatalf("Expect invalidArgument error")
	}
}

func TestStartContainer_invalidArgument_missingStartCommand(t *testing.T) {
	service := getService(&mockExecutor{})
	_, err := service.StartContainer(context.Background(), &api.StartContainerRequest{
		Name:           "my-container",
		ContainerImage: "us-docker.pkg.dev/cros-registry/test-services/cros-dut:8811903382633993457",
	})
	if err == nil {
		t.Fatalf("Expect invalidArgument error")
	}
}

func TestStartContainer_invalidPort_multiple(t *testing.T) {
	service := getService(&mockExecutor{})
	_, err := service.StartContainer(context.Background(), &api.StartContainerRequest{
		Name:              "my-container",
		ContainerImage:    "us-docker.pkg.dev/cros-registry/test-services/cros-dut:8811903382633993457",
		StartCommand:      []string{"cros-dut"},
		AdditionalOptions: &api.StartContainerRequest_Options{Expose: []string{"80", "90"}},
	})
	if err == nil {
		t.Fatalf("Expect unimpelemented error")
	}
}

func TestStartContainer_invalidPort_range(t *testing.T) {
	service := getService(&mockExecutor{})
	_, err := service.StartContainer(context.Background(), &api.StartContainerRequest{
		Name:              "my-container",
		ContainerImage:    "us-docker.pkg.dev/cros-registry/test-services/cros-dut:8811903382633993457",
		StartCommand:      []string{"cros-dut"},
		AdditionalOptions: &api.StartContainerRequest_Options{Expose: []string{"80-90"}},
	})
	if err == nil {
		t.Fatalf("Expect unimpelemented error")
	}
}

func TestStartContainer_pullError(t *testing.T) {
	errorMapping := make(map[string]string)
	errorMapping["*commands.DockerPull"] = "Permission \"artifactregistry.repositories.downloadArtifacts\" denied on resource"
	executor := mockExecutor{commandsToThrowError: errorMapping}
	service := getService(&executor)
	_, err := service.StartContainer(context.Background(), &api.StartContainerRequest{
		Name:           "my-container",
		ContainerImage: "us-docker.pkg.dev/cros-registry/test-services/cros-dut:8811903382633993457",
		StartCommand:   []string{"cros-dut"},
	})
	if err == nil {
		t.Fatalf("Expect permissionDenied error")
	}
	if len(executor.commandsExecuted) != 0 {
		t.Fatalf("Expect no commands have been executed")
	}
}

func TestStartContainer_runError(t *testing.T) {
	errorMapping := make(map[string]string)
	errorMapping["*commands.DockerRun"] = "Some unknown error"
	executor := mockExecutor{commandsToThrowError: errorMapping}
	service := getService(&executor)
	_, err := service.StartContainer(context.Background(), &api.StartContainerRequest{
		Name:           "my-container",
		ContainerImage: "us-docker.pkg.dev/cros-registry/test-services/cros-dut:8811903382633993457",
		StartCommand:   []string{"cros-dut"},
	})
	if err == nil {
		t.Fatalf("Expect unknown error")
	}
	if len(executor.commandsExecuted) != 1 {
		t.Fatalf("Expect 1 command has been executed")
	}
	if executor.commandsExecuted[0] != "*commands.DockerPull" {
		t.Fatalf("Expect docker pull have been executed")
	}
}

func TestStartContainer_success(t *testing.T) {
	executor := mockExecutor{}
	service := getService(&executor)
	_, err := service.StartContainer(context.Background(), &api.StartContainerRequest{
		Name:           "my-container",
		ContainerImage: "us-docker.pkg.dev/cros-registry/test-services/cros-dut:8811903382633993457",
		StartCommand:   []string{"cros-dut"},
	})
	if err != nil {
		t.Fatalf("Expect success")
	}
	if len(executor.commandsExecuted) != 2 {
		t.Fatalf("Expect 2 commands have been executed")
	}
	if executor.commandsExecuted[0] != "*commands.DockerPull" {
		t.Fatalf("Expect docker pull have been executed")
	}
	if executor.commandsExecuted[1] != "*commands.DockerRun" {
		t.Fatalf("Expect docker run have been executed")
	}
}

func TestStackCommands(t *testing.T) {
	executor := mockExecutor{}
	service := getService(&executor)
	_, err := service.StackCommands(context.Background(), &api.StackCommandsRequest{
		Requests: []*api.StackCommandsRequest_Stackable{
			{
				Command: &api.StackCommandsRequest_Stackable_CreateNetwork{
					CreateNetwork: &api.CreateNetworkRequest{Name: "bridge2"},
				},
			},
			{
				Command: &api.StackCommandsRequest_Stackable_CreateNetwork{
					CreateNetwork: &api.CreateNetworkRequest{Name: "bridge3"},
				},
			},
			{
				Command: &api.StackCommandsRequest_Stackable_StartContainer{
					StartContainer: &api.StartContainerRequest{
						Name:           "my-container",
						ContainerImage: "us-docker.pkg.dev/cros-registry/test-services/cros-dut:8811903382633993457",
						StartCommand:   []string{"cros-dut"},
					},
				},
			},
		}})
	if err != nil {
		t.Fatalf("Expect success")
	}
	if len(executor.commandsExecuted) != 4 {
		t.Fatalf("Expect 4 commands have been executed")
	}
	if executor.commandsExecuted[0] != "*commands.NetworkCreate" {
		t.Fatalf("Expect docker network create have been executed")
	}
	if executor.commandsExecuted[1] != "*commands.NetworkCreate" {
		t.Fatalf("Expect docker network create have been executed")
	}
	if executor.commandsExecuted[2] != "*commands.DockerPull" {
		t.Fatalf("Expect docker pull have been executed")
	}
	if executor.commandsExecuted[3] != "*commands.DockerRun" {
		t.Fatalf("Expect docker run have been executed")
	}
}
