// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"fmt"
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// ExecutorConfig represents executor configs.
type ExecutorConfig struct {
	InvServiceAddress string

	execsMap map[interfaces.ExecutorType]interfaces.ExecutorInterface
}

func NewExecutorConfig() interfaces.ExecutorConfigInterface {
	return &ExecutorConfig{execsMap: make(map[interfaces.ExecutorType]interfaces.ExecutorInterface)}
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

	default:
		return nil, fmt.Errorf("Executor type %s not supported in executor configs!", execType)
	}

	cfg.execsMap[execType] = exec
	return exec, nil
}
