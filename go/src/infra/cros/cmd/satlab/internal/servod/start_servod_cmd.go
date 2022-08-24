// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package servod

import (
	"context"
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/components/ufs"
	"infra/cros/cmd/satlab/internal/site"
	"infra/cros/recovery/docker"
	ufsApi "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// StartServodCmd is the command that will start a servod container
var StartServodCmd = &subcommands.Command{
	UsageLine: "start -host <hostname> [options ...]",
	ShortDesc: "starts servod container",
	LongDesc:  "Starts servod container",
	CommandRun: func() subcommands.CommandRun {
		c := &startServodRun{}

		c.Flags.StringVar(&c.host, "host", "", "Hostname of DUT")
		c.Flags.StringVar(&c.board, "board", "", "Board of DUT")
		c.Flags.StringVar(&c.model, "model", "", "Model of DUT")
		c.Flags.StringVar(&c.servoSerial, "servo-serial", "", "Servo Serial of DUT")
		c.Flags.StringVar(&c.servodContainerName, "servod-container-name", "", "Container name to run servod in; likely <host>-docker_servod")
		c.Flags.BoolVar(&c.noServodProcess, "no-servod", false, "Start container without the servod process running")

		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)

		return c
	},
}

// startServodRun struct contains the arguments for the servod command
type startServodRun struct {
	subcommands.CommandRunBase

	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	host string

	board               string
	model               string
	servoSerial         string
	servodContainerName string
	noServodProcess     bool
}

// Run is what is called when a user inputs the startServodRun command
func (c *startServodRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

// innerRun handles all orchestration needed to pass appropriate clients, contexts, and commands into the application
// this pattern is done to facilitate easy testing of the `runOrchestratedCommand` function
func (c *startServodRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	dhbSatlabID, err := commands.GetDockerHostBoxIdentifier()
	if err != nil {
		return err
	}

	if err := c.validate(dhbSatlabID, args); err != nil {
		cmdlib.PrintError(a, err)
		return err
	}

	ctx := cli.GetContext(a, c, env)
	ctx = utils.SetupContext(ctx, c.envFlags.Namespace)

	ef := c.envFlags
	ufs, err := ufs.NewUFSClient(ctx, ef.GetUFSService(), &c.authFlags)
	if err != nil {
		cmdlib.PrintError(a, errors.Reason("Error connecting to UFS: %s", err).Err())
		return err
	}

	d, err := docker.NewClient(ctx)
	if err != nil {
		cmdlib.PrintError(a, errors.Reason("Error creating docker client: %s", err).Err())
		return err
	}

	return c.runOrchestratedCommand(ctx, d, ufs)
}

// runOrchestratedCommand contains the actual logic behind this command. It does two things
// 	1. if the user does not provide all arguments needed to start the docker container, fetches from UFS
//  2. execute the docker command with the information either provided by the user or from UFS
func (c *startServodRun) runOrchestratedCommand(ctx context.Context, d DockerClient, ufs ufs.UFSClient) error {
	opts := ServodContainerOptions{
		containerName: c.servodContainerName,
		board:         c.board,
		model:         c.model,
		servoSerial:   c.servoSerial,
		withServod:    !c.noServodProcess, // notice negation here
	}

	// If user provides all needed data, we can skip the UFS fetch entirely which has utility for a DUT not deployed in UFS or when UFS is unreachable
	if err := opts.Validate(); err != nil {
		ufsMetadata, err := fetchMetadataFromUFS(ctx, ufs, c.host, &c.authFlags)

		if c.commonFlags.Verbose {
			fmt.Printf("Fetched metadata from UFS, recieved %+v", ufsMetadata)
		}

		if err != nil {
			return errors.Reason("Failed to fetch metadata from UFS. If all of -board, -model, -servo-serial, -servo-hostname are provided UFS fetch will be skipped: %v", err).Err()
		}

		// If partial data is provided (ex. subset of needed args) then we fetch all from UFS but only override variables not given
		if opts.board == "" {
			opts.board = ufsMetadata.board
		}
		if opts.model == "" {
			opts.model = ufsMetadata.model
		}
		if opts.servoSerial == "" {
			opts.servoSerial = ufsMetadata.servoSerial
		}
		if opts.containerName == "" {
			opts.containerName = ufsMetadata.servodContainerName
		}

		// If any field still is "", means that both user input and UFS fetch did not provide the appropriate field
		if err := opts.Validate(); err != nil {
			return errors.Reason("Fetch from UFS had empty fields, indicating data we need is not in UFS. Please specific args like -board, -model, manually").Err()
		}
	}

	dockerArgs := buildServodContainerArgs(opts)
	if c.commonFlags.Verbose {
		fmt.Printf("Attempting to launch container with command:\n\t%s\n", docker.StartCommandString(opts.containerName, dockerArgs))
	}
	_, err := startServodContainer(ctx, d, opts.containerName, dockerArgs)

	if err != nil {
		return errors.Reason(fmt.Sprintf("Error launching docker container: %s", err)).Err()
	}

	return nil
}

// fetchMetadataFromUFS pulls information about the asset and DUT of a given host
func fetchMetadataFromUFS(ctx context.Context, ufsClient ufs.UFSClient, host string, authFlags *authcli.Flags) (ufsMetadata, error) {
	dut, err := ufsClient.GetDut(ctx, &ufsApi.GetMachineLSERequest{
		Name: ufsUtil.AddPrefix(ufsUtil.MachineLSECollection, host),
	})
	if err != nil {
		return ufsMetadata{}, errors.Reason("Error fetching DUT %s from UFS: %s", host, err).Err()
	}

	servo := dut.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().GetServo()
	servoSerial := servo.GetServoSerial()
	servodContainerName := servo.GetDockerContainerName()

	if len(dut.GetMachines()) == 0 {
		return ufsMetadata{}, errors.Reason("Fetched DUT %s has no machineId", host).Err()
	}
	machineId := dut.GetMachines()[0]

	machine, err := ufsClient.GetMachine(ctx, &ufsApi.GetMachineRequest{
		Name: ufsUtil.AddPrefix(ufsUtil.MachineCollection, machineId),
	})
	if err != nil {
		return ufsMetadata{}, errors.Reason("Error fetching machine %s from UFS: %s", host, err).Err()
	}

	model := machine.GetChromeosMachine().GetModel()
	board := machine.GetChromeosMachine().GetBuildTarget()

	return ufsMetadata{board: board, model: model, servoSerial: servoSerial, servodContainerName: servodContainerName}, nil
}

// ufsMetadata is bag of data for fields we want to extract from UFS
type ufsMetadata struct {
	board               string
	model               string
	servoSerial         string
	servodContainerName string
}

// validate validates input arguments
// We primarily care that a) host is not empty and b) host is in the right format (satlab-<dhb_id>-<host>)
func (c *startServodRun) validate(dhbSatlabID string, positionalArgs []string) error {
	// ensures we did not recieve positional args
	if len(positionalArgs) > 0 {
		return errors.Reason("Got unexpected positional args, for usage see: satlab servo help start").Err()
	}

	// ensures the host in the startServoCommand is not empty
	if c.host == "" {
		return errors.Reason(fmt.Sprintf("-host <hostname> is required")).Err()
	}

	c.host = site.GetFullyQualifiedHostname(c.commonFlags.SatlabID, dhbSatlabID, site.Satlab, c.host)

	return nil
}

// hasEmptyString checks a list of strings for ""
// for readability
func hasEmptyString(args ...string) bool {
	for _, s := range args {
		if s == "" {
			return true
		}
	}

	return false
}
