// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"go.chromium.org/chromiumos/infra/proto/go/lab_platform"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_local_state"

	"infra/cros/cmd/phosphorus/internal/botcache"
	"infra/cros/cmd/phosphorus/internal/skylab_local_state/location"
	"infra/cros/cmd/phosphorus/internal/skylab_local_state/ufs"
	ufsutil "infra/unifiedfleet/app/util"
)

// Save subcommand: Update the bot state json file.
func Save(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "save -input_json /path/to/input.json",
		ShortDesc: "Update the DUT state json file.",
		LongDesc: `Update the DUT state json file.

(Re)Create the DUT state cache file using the state string from the input file
and provisionable labels and attributes from the host info file.
`,
		CommandRun: func() subcommands.CommandRun {
			c := &saveRun{}

			c.authFlags.Register(&c.Flags, authOpts)

			c.Flags.StringVar(&c.inputPath, "input_json", "", "Path to JSON SaveRequest to read.")
			return c
		},
	}
}

type saveRun struct {
	subcommands.CommandRunBase

	authFlags authcli.Flags

	inputPath string
}

func (c *saveRun) Run(a subcommands.Application, _ []string, env subcommands.Env) int {
	if err := c.validateArgs(); err != nil {
		fmt.Fprintln(a.GetErr(), err.Error())
		c.Flags.Usage()
		return 1
	}

	err := c.innerRun(a, env)
	if err != nil {
		fmt.Fprintln(a.GetErr(), err.Error())
		return 1
	}
	return 0
}

func (c *saveRun) validateArgs() error {
	if c.inputPath == "" {
		return fmt.Errorf("-input_json not specified")
	}

	return nil
}

func (c *saveRun) innerRun(a subcommands.Application, env subcommands.Env) error {
	var request skylab_local_state.SaveRequest
	if err := readJSONPb(c.inputPath, &request); err != nil {
		return err
	}

	if err := validateSaveRequest(&request); err != nil {
		return err
	}

	// Ensure save applies to all DUTs in a multi-DUTs task.
	targets := append(request.GetPeerDuts(), request.GetDutName())

	ctx := cli.GetContext(a, c, env)
	ctx = ufs.SetupContext(ctx, request.GetConfig().GetUfsNamespace())

	for _, hostname := range targets {
		bcs := botcache.Store{
			CacheDir: request.GetConfig().GetAutotestDir(),
			Name:     hostname,
		}
		logging.Infof(ctx, "Starting loading data from bot for %q...", hostname)
		s, err := bcs.Load()
		if err != nil {
			logging.Infof(ctx, "Fail to load data from bot: %v", err)
			return err
		}
		logging.Infof(ctx, "Starting loading the state file for %q...", hostname)
		i, err := getHostInfo(request.GetResultsDir(), hostname)
		if err != nil {
			logging.Infof(ctx, "Fail to load the state file: %v", err)
			return err
		}
		logging.Infof(ctx, "Starting updating request from the state file for %q...", hostname)
		s = updateDutStateFromHostInfo(s, i)
		logging.Infof(ctx, "Starting updating the state file for %q...", hostname)
		if err := bcs.Save(s); err != nil {
			logging.Infof(ctx, "Fail to update the state file: %v", err)
			return err
		}

		// Update the DUT state in UFS (if the current state is safe to update).
		if err := ufs.SafeUpdateUFSDUTState(ctx, &c.authFlags, hostname, request.GetDutState(), request.GetConfig().GetCrosUfsService(), request.GetRepairRequests()); err != nil {
			logging.Infof(ctx, "Fail to update state in UFS: %v", err)
		}
	}

	if request.GetSealResultsDir() {
		logging.Infof(ctx, "Sealing results dir is ignored since gsoffloader is not used anymore")
	}

	return nil
}

// validateSaveRequest checks for missing args and inserts any default values
func validateSaveRequest(request *skylab_local_state.SaveRequest) error {
	if request == nil {
		return fmt.Errorf("nil request")
	}

	var missingArgs []string

	if request.Config.GetAutotestDir() == "" {
		missingArgs = append(missingArgs, "autotest dir")
	}

	if request.ResultsDir == "" {
		missingArgs = append(missingArgs, "results dir")
	}

	if request.DutName == "" {
		missingArgs = append(missingArgs, "DUT hostname")
	}

	if request.DutId == "" {
		missingArgs = append(missingArgs, "DUT ID")
	}

	if request.DutState == "" {
		missingArgs = append(missingArgs, "DUT state")
	}

	if len(missingArgs) > 0 {
		return fmt.Errorf("no %s provided", strings.Join(missingArgs, ", "))
	}

	// note that this is adding a default value, not throwing an error
	if request.Config.GetUfsNamespace() == "" {
		request.Config.UfsNamespace = ufsutil.OSNamespace
	}

	return nil
}

// getHostInfo reads the host info from the store file.
func getHostInfo(resultsDir string, dutName string) (*skylab_local_state.AutotestHostInfo, error) {
	p := location.HostInfoFilePath(resultsDir, dutName)
	i := skylab_local_state.AutotestHostInfo{}

	if err := readJSONPb(p, &i); err != nil {
		return nil, errors.Annotate(err, "get host info for %s", dutName).Err()
	}

	return &i, nil
}

var provisionableAttributes = map[string]bool{
	"job_repo_url":   true,
	"outlet_changed": true,
}

func updateDutStateFromHostInfo(s *lab_platform.DutState, i *skylab_local_state.AutotestHostInfo) *lab_platform.DutState {
	as := make(map[string]string)
	for attribute, value := range i.GetAttributes() {
		if provisionableAttributes[attribute] {
			as[attribute] = value
		}
	}
	s.ProvisionableAttributes = updateMap(s.ProvisionableAttributes, as)
	return s
}

func updateMap(to, from map[string]string) map[string]string {
	if len(from) == 0 {
		return to
	}
	if to == nil {
		to = make(map[string]string)
	}
	for k, v := range from {
		to[k] = v
	}
	return to
}

const gsOffloaderMarker = ".ready_for_offload"

// sealResultsDir drops a special timestamp file in the results directory
// notifying gs_offloader to offload the directory. The results directory
// should not be touched once sealed. This should not be called on an
// already sealed results directory.
// NOTE: NOT BEING USED CURRENTLY SINCE GSOFFLOADER IS NOT USED ANYMORE.
func sealResultsDir(dir string) error {
	ts := []byte(fmt.Sprintf("%d", time.Now().Unix()))
	tsFile := filepath.Join(dir, gsOffloaderMarker)
	_, err := os.Stat(tsFile)
	if err == nil {
		return fmt.Errorf("seal results dir %s: already sealed", dir)
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("seal results dir %s: encountered corrupted file %s", dir, tsFile)
	}
	if err := ioutil.WriteFile(tsFile, ts, 0666); err != nil {
		return errors.Annotate(err, "seal results dir %s", dir).Err()
	}
	return nil
}
