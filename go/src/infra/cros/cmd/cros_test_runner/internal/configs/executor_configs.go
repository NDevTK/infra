// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/containers"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"

	"go.chromium.org/luci/common/errors"
)

// ExecutorConfig represents executor configs.
type ExecutorConfig struct {
	ContainerConfig   interfaces.ContainerConfigInterface
	InvServiceAddress string
	Ctr               *crostoolrunner.CrosToolRunner

	execsMap map[interfaces.ExecutorType]interfaces.ExecutorInterface
}

func NewExecutorConfig(
	ctr *crostoolrunner.CrosToolRunner,
	contConfig interfaces.ContainerConfigInterface) interfaces.ExecutorConfigInterface {
	execsMap := make(map[interfaces.ExecutorType]interfaces.ExecutorInterface)
	return &ExecutorConfig{Ctr: ctr, ContainerConfig: contConfig, execsMap: execsMap}
}

// GetExecutor returns the concrete executor based on provided executor type.
func (cfg *ExecutorConfig) GetExecutor(execType interfaces.ExecutorType) (interfaces.ExecutorInterface, error) {
	// Return executor if already created.
	if savedExec, ok := cfg.execsMap[execType]; ok {
		return savedExec, nil
	}

	var exec interfaces.ExecutorInterface

	// Get executor type based on executor type.
	switch execType {
	case executors.InvServiceExecutorType:
		invServiceAddress := ""
		if cfg.InvServiceAddress != "" {
			invServiceAddress = cfg.InvServiceAddress
		}
		exec = executors.NewInvServiceExecutor(invServiceAddress)

	case executors.CtrExecutorType:
		if cfg.Ctr == nil {
			return nil, fmt.Errorf("CrosToolRunner is nil!")
		}
		exec = executors.NewCtrExecutor(cfg.Ctr)

	case executors.CrosDutExecutorType:
		container, err := cfg.ContainerConfig.GetContainer(containers.CrosDutTemplatedContainerType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container for executor type %s", execType).Err()
		}
		exec = executors.NewCrosDutExecutor(container)

	case executors.CrosProvisionExecutorType:
		container, err := cfg.ContainerConfig.GetContainer(containers.CrosProvisionTemplatedContainerType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container for executor type %s", execType).Err()
		}
		exec = executors.NewCrosProvisionExecutor(container)

	case executors.CrosTestExecutorType:
		container, err := cfg.ContainerConfig.GetContainer(containers.CrosTestTemplatedContainerType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container for executor type %s", execType).Err()
		}
		exec = executors.NewCrosTestExecutor(container)

	case executors.CrosGcsPublishExecutorType:
		container, err := cfg.ContainerConfig.GetContainer(containers.CrosGcsPublishTemplatedContainerType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container for executor type %s", execType).Err()
		}
		exec = executors.NewCrosPublishExecutor(container, execType)

	case executors.CrosTkoPublishExecutorType:
		container, err := cfg.ContainerConfig.GetContainer(containers.CrosTkoPublishTemplatedContainerType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container for executor type %s", execType).Err()
		}
		exec = executors.NewCrosPublishExecutor(container, execType)

	case executors.CrosRdbPublishExecutorType:
		container, err := cfg.ContainerConfig.GetContainer(containers.CrosRdbPublishTemplatedContainerType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container for executor type %s", execType).Err()
		}
		exec = executors.NewCrosPublishExecutor(container, execType)

	case executors.SshTunnelExecutorType:
		exec = executors.NewSshTunnelExecutor()

	case executors.CacheServerExecutorType:
		container, err := cfg.ContainerConfig.GetContainer(containers.CacheServerTemplatedContainerType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container for executor type %s", execType).Err()
		}
		exec = executors.NewCacheServerExecutor(container)

	case executors.CrosTestFinderExecutorType:
		container, err := cfg.ContainerConfig.GetContainer(containers.CrosTestFinderTemplatedContainerType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container for executor type %s", execType).Err()
		}
		exec = executors.NewCrosTestFinderExecutor(container)

	default:
		return nil, fmt.Errorf("Executor type %s not supported in executor configs!", execType)
	}

	cfg.execsMap[execType] = exec
	return exec, nil
}
