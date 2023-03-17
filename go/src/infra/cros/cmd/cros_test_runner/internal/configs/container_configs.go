// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"fmt"

	"infra/cros/cmd/cros_test_runner/common"
	"infra/cros/cmd/cros_test_runner/internal/containers"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"

	"go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/luci/common/errors"
)

// CftContainerConfig represents Cft container configs.
type CftContainerConfig struct {
	Ctr                *crostoolrunner.CrosToolRunner
	ContainerImagesMap map[string]*api.ContainerImageInfo

	containersMap map[interfaces.ContainerType]interfaces.ContainerInterface
}

func NewCftContainerConfig(
	ctr *crostoolrunner.CrosToolRunner,
	containerImagesMap map[string]*api.ContainerImageInfo) interfaces.ContainerConfigInterface {

	contMap := make(map[interfaces.ContainerType]interfaces.ContainerInterface)
	return &CftContainerConfig{
		Ctr:                ctr,
		ContainerImagesMap: containerImagesMap,
		containersMap:      contMap,
	}
}

// GetContainer returns the concrete container based on provided container type.
func (cfg *CftContainerConfig) GetContainer(contType interfaces.ContainerType) (interfaces.ContainerInterface, error) {
	// Return container if already created.
	if savedCont, ok := cfg.containersMap[contType]; ok {
		return savedCont, nil
	}

	if len(cfg.ContainerImagesMap) == 0 {
		return nil, fmt.Errorf("ContainerImagesMap is empty!")
	}
	if cfg.Ctr == nil {
		return nil, fmt.Errorf("CrosToolRunner is nil!")
	}

	var cont interfaces.ContainerInterface
	key := containers.GetContainerImageKeyFromContainerType(contType)
	// Get container based on container type.
	switch contType {
	case containers.CrosDutTemplatedContainerType:
		containerImage, err := common.GetContainerImageFromMap(key, cfg.ContainerImagesMap)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container image from map for %s container type", contType).Err()
		}
		cont = containers.NewCrosDutTemplatedContainer(containerImage, cfg.Ctr)

	case containers.CrosProvisionTemplatedContainerType:
		containerImage, err := common.GetContainerImageFromMap(key, cfg.ContainerImagesMap)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container image from map for %s container type", contType).Err()
		}
		cont = containers.NewCrosProvisionTemplatedContainer(containerImage, cfg.Ctr)

	case containers.CrosTestTemplatedContainerType:
		containerImage, err := common.GetContainerImageFromMap(key, cfg.ContainerImagesMap)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container image from map for %s container type", contType).Err()
		}
		cont = containers.NewCrosTestTemplatedContainer(containerImage, cfg.Ctr)

	case containers.CrosTestFinderTemplatedContainerType:
		containerImage, err := common.GetContainerImageFromMap(key, cfg.ContainerImagesMap)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container image from map for %s container type", contType).Err()
		}
		cont = containers.NewCrosTestFinderTemplatedContainer(containerImage, cfg.Ctr)

	case containers.CrosGcsPublishTemplatedContainerType, containers.CrosTkoPublishTemplatedContainerType, containers.CrosRdbPublishTemplatedContainerType, containers.CrosCpconPublishTemplatedContainerType:
		containerImage, err := common.GetContainerImageFromMap(key, cfg.ContainerImagesMap)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container image from map for %s container type", contType).Err()
		}
		cont = containers.NewCrosPublishTemplatedContainer(contType, containerImage, cfg.Ctr)

	case containers.CacheServerTemplatedContainerType:
		containerImage := common.DockerImageCacheServer
		cont = containers.NewCacheServerTemplatedContainer(containerImage, cfg.Ctr)

	default:
		return nil, fmt.Errorf("Container type %s not supported in container configs!", contType)
	}

	cfg.containersMap[contType] = cont
	return cont, nil
}
