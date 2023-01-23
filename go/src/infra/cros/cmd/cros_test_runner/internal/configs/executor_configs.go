// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
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
	return nil, nil
}
