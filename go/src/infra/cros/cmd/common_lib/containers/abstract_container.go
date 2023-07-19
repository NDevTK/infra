// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package containers

import (
	"context"
	"fmt"
	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"strings"

	"github.com/google/uuid"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// Container state types
type ContainerState string

const (
	ContainerStateNotInitialized ContainerState = "NotInitialized"
	ContainerStateInitialized    ContainerState = "Initialized"
	ContainerStateStarted        ContainerState = "Started"
	ContainerStateStopped        ContainerState = "Stopped"
)

func NewContainer(
	contType interfaces.ContainerType,
	namePrefix string,
	containerImage string,
	ctr *crostoolrunner.CrosToolRunner,
	isTemplated bool) interfaces.ContainerInterface {

	if isTemplated {
		return NewTemplatedContainer(contType, namePrefix, containerImage, ctr)
	} else {
		return NewNonTemplatedContainer(contType, namePrefix, containerImage, ctr)
	}
}

// AbstractContainer represents abstract container.
type AbstractContainer struct {
	interfaces.ContainerInterface

	namePrefix     string
	containerImage string
	ctr            *crostoolrunner.CrosToolRunner
	state          ContainerState

	Name          string
	TempDirLoc    string
	containerType interfaces.ContainerType

	ConcreteContainer  interfaces.ContainerInterface
	StartContainerResp *api.StartContainerResponse
	GetContainerResp   *api.GetContainerResponse
}

func NewAbstractContainer(
	contType interfaces.ContainerType,
	namePrefix string,
	containerImage string,
	ctr *crostoolrunner.CrosToolRunner) AbstractContainer {

	return AbstractContainer{containerType: contType, namePrefix: namePrefix, containerImage: containerImage, ctr: ctr, state: ContainerStateNotInitialized}
}

func (cont *AbstractContainer) GetContainerType() interfaces.ContainerType {
	return cont.containerType
}

func (cont *AbstractContainer) GetLogsLocation() (string, error) {
	if cont.TempDirLoc == "" {
		return "", fmt.Errorf("Temp dir is not created yet for %s container!", cont.GetContainerType())
	}

	return cont.TempDirLoc, nil
}

// InitializeBase does initial work that is common to all containers.
func (cont *AbstractContainer) InitializeBase(ctx context.Context) error {
	if cont.state != ContainerStateNotInitialized && cont.state != ContainerStateStopped {
		return fmt.Errorf(
			"Expected state %s or %s during initializing, found state %s instead!",
			ContainerStateNotInitialized,
			ContainerStateStopped,
			cont.state)
	}
	if cont.namePrefix == "" {
		return fmt.Errorf("No name prefix provided for container")
	}
	if cont.containerImage == "" {
		return fmt.Errorf("No container image provided for cros-test container")
	}

	id := uuid.New().String()
	cont.Name = fmt.Sprintf("%s-container-%s", cont.namePrefix, strings.Split(id, "-")[0])

	tempDirLoc, err := common.CreateTempDir(ctx, cont.namePrefix)
	if err != nil {
		return errors.Annotate(err, "Failed to create temp dir for %s", cont.namePrefix).Err()
	}

	cont.TempDirLoc = tempDirLoc
	logging.Infof(ctx, fmt.Sprintf("Temp dir created for %s: %s", cont.namePrefix, tempDirLoc))

	return nil
}

// GetContainer gets the container info.
func (cont *AbstractContainer) GetContainer(ctx context.Context) (*api.GetContainerResponse, error) {
	if cont.state != ContainerStateStarted {
		return nil, fmt.Errorf(
			"Expected state %s during getting container, found state %s instead!",
			ContainerStateStarted,
			cont.state)
	}
	if cont.Name == "" {
		return nil, fmt.Errorf("Container name not found while trying to get the container!")
	}

	var err error
	cont.GetContainerResp, err = cont.ctr.GetContainer(ctx, cont.Name)
	if err != nil {
		return nil, errors.Annotate(err, "error getting container %s", cont.Name).Err()
	}
	return cont.GetContainerResp, err
}

// StopContainer stop the container.
func (cont *AbstractContainer) StopContainer(ctx context.Context) error {
	if cont.state != ContainerStateStarted {
		return fmt.Errorf(
			"Expected state %s during stopping container, found state %s instead!",
			ContainerStateStarted,
			cont.state)
	}

	if cont.Name == "" {
		return fmt.Errorf("Container name not found while trying to get the container!")
	}

	var err error
	err = cont.ctr.StopContainer(ctx, cont.Name)
	if err != nil {
		return errors.Annotate(err, "error getting container %s", cont.Name).Err()
	}

	cont.state = ContainerStateStopped

	return err
}

// ProcessContainer processes(initialize, start, get, retrieve server address)
// the container.
func (cont *AbstractContainer) ProcessContainer(
	ctx context.Context,
	template *api.Template) (string, error) {

	if err := cont.ConcreteContainer.Initialize(ctx, template); err != nil {
		return "", errors.Annotate(err, "error during initializing container: ").Err()
	}

	if _, err := cont.ConcreteContainer.StartContainer(ctx); err != nil {
		return "", errors.Annotate(err, "error during starting container: ").Err()
	}

	getContResp, err := cont.ConcreteContainer.GetContainer(ctx)
	if err != nil {
		return "", errors.Annotate(err, "error during getting container: ").Err()
	}

	serverAddress, err := common.GetServerAddressFromGetContResponse(getContResp)
	if err != nil {
		return "", errors.Annotate(err, "error getting server address: ").Err()
	}

	return serverAddress, nil
}
