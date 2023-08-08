// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/mallet/internal/site"
	"infra/cmdsupport/cmdlib"
)

// mallet test-shivas runs an integration test on shivas.
//
// It uses the `./dev-shivas` (which is `shivas` built with a dev tag) to manipulate
// a UFS instance running locally.
//
// The `dev-shivas` target is incapable of manipulating prod data, so this tool, too, should
// only be capable of talking to the dev shivas.

// Perform an integration test on shivas
var TestShivas = &subcommands.Command{
	UsageLine: "test-shivas",
	ShortDesc: "Test shivas CLI",
	LongDesc:  "Test shivas CLI",
	CommandRun: func() subcommands.CommandRun {
		c := &testShivasRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.Flags.StringVar(&c.dir, "dir", "", `directory where shivas command is located`)
		return c
	},
}

type testShivasRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags

	dir string
}

func (c *testShivasRun) devShivas() string {
	if c.dir == "" {
		return ""
	}
	return filepath.Join(c.dir, "dev-shivas")
}

func (c *testShivasRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *testShivasRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if len(args) != 0 {
		return errors.Reason("test shivas: positional arguments are not supported").Err()
	}
	if c.dir == "" {
		return errors.Reason("test shivas: argument -dir must be provided").Err()
	}

	if err := exec.Command(c.devShivas(), "help").Run(); err != nil {
		return errors.Annotate(err, `test shivas: "dev-shivas help" failed`).Err()
	}
	if err := c.addRemoveDUT(); err != nil {
		return errors.Annotate(err, "test shivas: add remove dut").Err()
	}
	if err := c.auditServoUSB(); err != nil {
		return errors.Annotate(err, "test shivas: audit servo usb").Err()
	}

	fmt.Fprintf(a.GetErr(), "OK\n")
	return nil
}

// addRemoveDUT tests adding a DUT and then removing it.
//
// In order to test this interaction, we first add a rack, then an asset, then a DUT.
// At the end we tear everything down in the opposite order.
//
// Keep this test simple. This is supposed to be a model test that other tests will be based on.
func (c *testShivasRun) addRemoveDUT() error {
	const zone = "cros_googler_desk"
	const rack = "937d6144-9a54-4967-a31b-9c134a823150"
	const asset = "3fab83ca-1295-454b-9881-7d0e8b37fbe5"
	const dut = "49e48288-4826-4074-a10a-99e1858ad78e"
	const eve = "eve"

	f := func(e error) error {
		if e == nil {
			return errors.Reason("command exited zero but another problem was detected").Err()
		}
		return e
	}

	cleanup := func() error {
		var merr errors.MultiError
		merr.MaybeAdd(exec.Command(c.devShivas(), "delete", "dut", "-yes", "-namespace", "os", dut).Run())
		merr.MaybeAdd(exec.Command(c.devShivas(), "delete", "asset", "-yes", "-namespace", "os", asset).Run())
		merr.MaybeAdd(exec.Command(c.devShivas(), "delete", "rack", "-yes", "-namespace", "os", rack).Run())
		return merr.AsError()
	}

	// intentionally ignore the errors during initial cleanup and final cleanup.
	cleanup()
	defer cleanup()

	if out, err := exec.Command(c.devShivas(), "add", "rack", "-namespace", "os", "-zone", zone, "-name", rack).CombinedOutput(); !strings.Contains(string(out), "Success") || err != nil {
		return errors.Annotate(f(err), "add remove DUT: add rack: %q", out).Err()
	}
	if out, err := exec.Command(c.devShivas(), "add", "asset", "-namespace", "os", "-zone", zone, "-type", "dut", "-rack", rack, "-name", asset).CombinedOutput(); !strings.Contains(string(out), "Success") || err != nil {
		return errors.Annotate(f(err), "add remove DUT: add asset: %q", out).Err()
	}
	if out, err := exec.Command(c.devShivas(), "add", "dut", "-namespace", "os", "-asset", asset, "-board", eve, "-model", eve, "-name", dut).CombinedOutput(); !strings.Contains(string(out), "Success") || err != nil {
		return errors.Annotate(f(err), "add remove DUT: add dut: %q", out).Err()
	}

	// Try to clean up and track whether we succeeded or failed.
	return errors.Annotate(cleanup(), "add dut: cleanup").Err()
}

// auditServoUSB tests launching an audit-servo-usb task.
func (c *testShivasRun) auditServoUSB() error {
	const zone = "cros_googler_desk"
	const rack = "ee49b4de-c6e5-4790-a0d3-b9e184666386"
	const asset = "d7c3d309-9ce3-4dfe-9812-9683c2b22b87"
	const dut = "674a87bd-c6e9-4dad-ae51-b08ddd48677d"
	const eve = "eve"

	f := func(e error) error {
		if e == nil {
			return errors.Reason("command exited zero but another problem was detected").Err()
		}
		return e
	}

	cleanup := func() {
		exec.Command(c.devShivas(), "delete", "dut", "-yes", "-namespace", "os", dut).Run()
		exec.Command(c.devShivas(), "delete", "asset", "-yes", "-namespace", "os", asset).Run()
		exec.Command(c.devShivas(), "delete", "rack", "-yes", "-namespace", "os", rack).Run()
	}

	cleanup()
	defer cleanup()

	// Create the DUT.
	if out, err := exec.Command(c.devShivas(), "add", "rack", "-namespace", "os", "-zone", zone, "-name", rack).CombinedOutput(); !strings.Contains(string(out), "Success") || err != nil {
		return errors.Annotate(f(err), "add remove DUT: add rack: %q", out).Err()
	}
	if out, err := exec.Command(c.devShivas(), "add", "asset", "-namespace", "os", "-zone", zone, "-type", "dut", "-rack", rack, "-name", asset).CombinedOutput(); !strings.Contains(string(out), "Success") || err != nil {
		return errors.Annotate(f(err), "add remove DUT: add asset: %q", out).Err()
	}
	if out, err := exec.Command(c.devShivas(), "add", "dut", "-namespace", "os", "-asset", asset, "-board", eve, "-model", eve, "-name", dut).CombinedOutput(); !strings.Contains(string(out), "Success") || err != nil {
		return errors.Annotate(f(err), "add remove DUT: add dut: %q", out).Err()
	}

	// Launch an audit task.
	// See b:244778016 for more details.
	//
	// TODO(gregorynisbet): For now, we expect audit tasks launched against the dev version of UFS to fail with an error containing
	//                      "Method Not Allowed". In future CLs, fix this so that we actually launch a swarming task in the dev instance.
	if out, err := exec.Command(c.devShivas(), "audit", "-servo-usb", dut).CombinedOutput(); !strings.Contains(string(out), "Method Not Allowed") {
		return errors.Annotate(f(err), "audit DUT: launch audit task: %q", out).Err()
	}

	return nil
}
