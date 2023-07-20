// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"fmt"

	"infra/cros/cmd/common_lib/common_executors"
	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/ctpv2/internal/executors"

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
	case common_executors.CtrExecutorType:
		if cfg.Ctr == nil {
			return nil, fmt.Errorf("CrosToolRunner is nil!")
		}
		exec = common_executors.NewCtrExecutor(cfg.Ctr)

	case executors.FilterExecutorType:
		container, err := cfg.ContainerConfig.GetContainer(containers.CrosTestFinderTemplatedContainerType)
		if err != nil {
			return nil, errors.Annotate(err, "error during getting container for executor type %s", execType).Err()
		}
		exec = executors.NewFilterExecutor(container)

	default:
		return nil, fmt.Errorf("Executor type %s not supported in executor configs!", execType)
	}

	cfg.execsMap[execType] = exec
	return exec, nil
}
