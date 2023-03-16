// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package containers

import (
	"context"
	"fmt"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
)

// TemplatedContainer represents the cft templated container.
type TemplatedContainer struct {
	AbstractContainer

	StartTemplatedContainerReq *api.StartTemplatedContainerRequest
}

func NewTemplatedContainer(contType interfaces.ContainerType,
	namePrefix string,
	containerImage string,
	ctr *crostoolrunner.CrosToolRunner) *TemplatedContainer {

	cont := &TemplatedContainer{AbstractContainer: NewAbstractContainer(contType, namePrefix, containerImage, ctr)}
	cont.ConcreteContainer = cont
	return cont
}

// Initialize initializes the container.
func (cont *TemplatedContainer) Initialize(
	ctx context.Context,
	template *api.Template) error {

	err := cont.AbstractContainer.InitializeBase(ctx)
	if err != nil {
		return errors.Annotate(err, "initialization failed for base container: ").Err()
	}

	if template == nil {
		return fmt.Errorf("No template provided for templated container!")
	}

	switch t := template.Container.(type) {
	case *api.Template_CrosDut:
		if err = cont.initializeCrosDutTemplate(ctx, t.CrosDut); err != nil {
			return errors.Annotate(err, "initialization failed for cros-dut template: ").Err()
		}
	case *api.Template_CrosProvision:
		if err = cont.initializeCrosProvisionTemplate(ctx, t.CrosProvision); err != nil {
			return errors.Annotate(err, "initialization failed for cros-provision template: ").Err()
		}
	case *api.Template_CrosTest:
		if err = cont.initializeCrosTestTemplate(ctx, t.CrosTest); err != nil {
			return errors.Annotate(err, "initialization failed for cros-test template: ").Err()
		}
	case *api.Template_CrosTestFinder:
		if err = cont.initializeCrosTestFinderTemplate(ctx, t.CrosTestFinder); err != nil {
			return errors.Annotate(err, "initialization failed for cros-test-finder template: ").Err()
		}
	case *api.Template_CrosPublish:
		if err = cont.initializeCrosPublishTemplate(ctx, t.CrosPublish); err != nil {
			return errors.Annotate(err, "initialization failed for cros-publish template: ").Err()
		}
	case *api.Template_CacheServer:
		if err = cont.initializeCacheServerTemplate(ctx, t.CacheServer); err != nil {
			return errors.Annotate(err, "initialization failed for cache-server template: ").Err()
		}
	default:
		return fmt.Errorf("Provided template %v not found!", t)
	}

	if cont.TempDirLoc == "" {
		return fmt.Errorf("TempDirLoc is empty but required for ArtifactDir")
	}

	cont.StartTemplatedContainerReq = &api.StartTemplatedContainerRequest{
		Name:           cont.Name,
		ContainerImage: cont.containerImage,
		Template:       template,
		Network:        common.ContainerDefaultNetwork,
		ArtifactDir:    cont.TempDirLoc}

	cont.state = ContainerStateInitialized

	return nil
}

// initializeCrosDutTemplate initializes cros dut template.
func (cont *TemplatedContainer) initializeCrosDutTemplate(
	ctx context.Context,
	dutTemplate *api.CrosDutTemplate) error {

	if dutTemplate == nil {
		return fmt.Errorf("Provided CrosDutTemplate is nil!")
	}

	if dutTemplate.GetCacheServer() == nil {
		return fmt.Errorf("No cache server provided for dut template!")
	}

	if dutTemplate.GetDutAddress() == nil {
		return fmt.Errorf("No dut address provided for dut template")
	}

	return nil
}

// initializeCrosProvisionTemplate initializes cros provision template.
func (cont *TemplatedContainer) initializeCrosProvisionTemplate(
	ctx context.Context,
	provisionTemplate *api.CrosProvisionTemplate) error {

	if provisionTemplate == nil {
		return fmt.Errorf("Provided CrosProvisionTemplate is nil!")
	}

	if provisionTemplate.GetInputRequest() == nil {
		return fmt.Errorf("No input request provided for provision template!")
	}

	return nil
}

// initializeCrosTestTemplate initializes cros test template.
func (cont *TemplatedContainer) initializeCrosTestTemplate(
	ctx context.Context,
	testTemplate *api.CrosTestTemplate) error {

	if testTemplate == nil {
		return fmt.Errorf("Provided CrosTestTemplate is nil!")
	}

	return nil
}

// initializeCrosTestFinderTemplate initializes cros test finder template.
func (cont *TemplatedContainer) initializeCrosTestFinderTemplate(
	ctx context.Context,
	testFinderTemplate *api.CrosTestFinderTemplate) error {

	if testFinderTemplate == nil {
		return fmt.Errorf("Provided CrosTestFinderTemplate is nil!")
	}

	return nil
}

// initializeCacheServerTemplate initializes cache server template.
func (cont *TemplatedContainer) initializeCacheServerTemplate(
	ctx context.Context,
	cacheTemplate *api.CacheServerTemplate) error {

	if cacheTemplate == nil {
		return fmt.Errorf("Provided CacheServerTemplate is nil!")
	}

	return nil
}

// initializeCrosPublishTemplate initializes cros publish template.
func (cont *TemplatedContainer) initializeCrosPublishTemplate(
	ctx context.Context,
	publishTemplate *api.CrosPublishTemplate) error {

	if publishTemplate == nil {
		return fmt.Errorf("Provided CrosPublishTemplate is nil!")
	}

	if publishTemplate.PublishType == api.CrosPublishTemplate_PUBLISH_GCS ||
		publishTemplate.PublishType == api.CrosPublishTemplate_PUBLISH_TKO ||
		publishTemplate.PublishType == api.CrosPublishTemplate_PUBLISH_CPCON {
		if publishTemplate.PublishSrcDir == "" {
			return fmt.Errorf("PublishSrcDir is empty but required for GCS, TKO, and CPCON publish types!")
		}
	}

	return nil
}

// StartContainer starts the container.
func (cont *TemplatedContainer) StartContainer(ctx context.Context) (*api.StartContainerResponse, error) {
	if cont.StartTemplatedContainerReq == nil {
		return nil, fmt.Errorf("StartTemplatedContainerRequest is nil!")
	}
	if cont.ctr == nil {
		return nil, fmt.Errorf("Ctr is nil!")
	}
	var err error
	cont.StartContainerResp, err = cont.ctr.StartTemplatedContainer(ctx, cont.StartTemplatedContainerReq)
	if err != nil {
		return nil, errors.Annotate(err, "error starting templated container: ").Err()
	}

	cont.state = ContainerStateStarted

	return cont.StartContainerResp, nil
}
