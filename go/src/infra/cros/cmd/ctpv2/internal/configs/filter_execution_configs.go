// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/common_configs"
	"infra/cros/cmd/common_lib/interfaces"
)

// Config types
const (
	LuciBuildFilterExecutionConfigType interfaces.ConfigType = "LuciBuild"

	// For unit tests purposes only
	UnSupportedFilterExecutionConfigType interfaces.ConfigType = "UnsupportedTest"
)

// TestExecutionConfig represents the configuration for any test execution.
type Ctpv2ExecutionConfig struct {
	*common_configs.CmdExecutionConfig

	executedCommands map[interfaces.CommandType]bool
}

func NewCtpv2ExecutionConfig(
	configType interfaces.ConfigType,
	cmdConfig interfaces.CommandConfigInterface,
	ski interfaces.StateKeeperInterface) *Ctpv2ExecutionConfig {

	cmdExecutionConfig := common_configs.NewCmdExecutionConfig(configType, cmdConfig, ski)
	return &Ctpv2ExecutionConfig{
		CmdExecutionConfig: cmdExecutionConfig,
	}
}

func (ctpv2cfg *Ctpv2ExecutionConfig) GenerateConfig(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, fmt.Sprintf("Generate configs: %s", ctpv2cfg.GetConfigType()))
	defer func() { step.End(err) }()

	switch configType := ctpv2cfg.GetConfigType(); configType {
	case LuciBuildFilterExecutionConfigType:
		ctpv2cfg.Configs = GenerateFilterConfigs(ctx)
	default:
		err = fmt.Errorf("Config type %s is not supported!", configType)
	}

	if ctpv2cfg.Configs != nil {
		configsLog := step.Log("generated configs")
		_, logErr := configsLog.Write([]byte(ctpv2cfg.Configs.ToString()))
		if logErr != nil {
			logging.Infof(ctx, "error during writing generated configs: %s", logErr)
		}
	}

	return err
}
