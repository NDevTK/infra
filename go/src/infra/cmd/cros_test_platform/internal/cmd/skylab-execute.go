// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"

	"github.com/maruel/subcommands"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/isolatedclient"

	"infra/cmd/cros_test_platform/internal/execution"
	"infra/cmd/cros_test_platform/internal/execution/isolate"
	"infra/cmd/cros_test_platform/internal/execution/isolate/getter"
)

// SkylabExecute subcommand: Run a set of enumerated tests against skylab backend.
var SkylabExecute = &subcommands.Command{
	UsageLine: "skylab-execute -input_json /path/to/input.json -output_json /path/to/output.json",
	ShortDesc: "Run a set of enumerated tests against skylab backend.",
	LongDesc: `Run a set of enumerated tests against skylab backend.

	Placeholder only, not yet implemented.`,
	CommandRun: func() subcommands.CommandRun {
		c := &skylabExecuteRun{}
		c.addFlags()
		return c
	},
}

type skylabExecuteRun struct {
	commonExecuteRun
}

func (c *skylabExecuteRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.validateArgs(); err != nil {
		fmt.Fprintln(a.GetErr(), err.Error())
		c.Flags.Usage()
		return exitCode(err)
	}

	err := c.innerRun(a, args, env)
	if err != nil {
		fmt.Fprintf(a.GetErr(), "%s\n", err)
	}
	return exitCode(err)
}

func (c *skylabExecuteRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	request, err := c.readRequest(c.inputPath)
	if err != nil {
		return err
	}

	if err = c.validateRequest(request); err != nil {
		return err
	}

	ctx := cli.GetContext(a, c, env)
	ctx = setupLogging(ctx)

	client, err := swarmingClient(ctx, request.Config.SkylabSwarming)
	if err != nil {
		return err
	}

	gf := c.getterFactory(request.Config.SkylabIsolate)

	runner := execution.NewSkylabRunner(request.Enumeration.AutotestTests, request.RequestParams, request.Config.SkylabWorker)

	response, err := c.handleRequest(ctx, runner, client, gf)
	if err != nil && response == nil {
		// Catastrophic error. There is no reasonable response to write.
		return err
	}

	return writeResponse(c.outputPath, response, err)
}

func (c *skylabExecuteRun) validateRequest(request *steps.ExecuteRequest) error {
	if err := c.validateRequestCommon(request); err != nil {
		return err
	}

	if request.Config.SkylabSwarming == nil {
		return fmt.Errorf("nil request.config.skylab_swarming")
	}

	if request.Config.SkylabIsolate == nil {
		return fmt.Errorf("nil request.config.skylab_isolate")
	}

	if request.Config.SkylabWorker == nil {
		return fmt.Errorf("nil request.config.skylab_worker")
	}

	return nil
}

func (c *skylabExecuteRun) getterFactory(conf *config.Config_Isolate) isolate.GetterFactory {
	return func(ctx context.Context, server string) (isolate.Getter, error) {
		hClient, err := httpClient(ctx, conf.AuthJsonPath)
		if err != nil {
			return nil, err
		}

		isolateClient := isolatedclient.New(nil, hClient, server, isolatedclient.DefaultNamespace, nil, nil)

		return getter.New(isolateClient), nil
	}
}
