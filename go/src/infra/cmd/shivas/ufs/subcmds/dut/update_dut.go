// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/grpc/prpc"
	"google.golang.org/genproto/protobuf/field_mask"

	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	swarming "infra/libs/swarming"
	ufspb "infra/unifiedfleet/api/v1/models"
	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

const (
	// Servo related UpdateMask paths.
	servoHostPath   = "dut.servo.hostname"
	servoPortPath   = "dut.servo.port"
	servoSerialPath = "dut.servo.serial"
	servoSetupPath  = "dut.servo.setup"
)

var defaultRedeployTaskActions = []string{"run-pre-deploy-verification"}

// UpdateDUTCmd update dut by given hostname and start a swarming job to delpoy.
var UpdateDUTCmd = &subcommands.Command{
	UsageLine: "dut [options]",
	ShortDesc: "Update a DUT.",
	LongDesc:  cmdhelp.UpdateDUTLongDesc,
	CommandRun: func() subcommands.CommandRun {
		c := &updateDUT{
			deployTags:    shivasTags,
			deployActions: defaultRedeployTaskActions,
		}
		// Initialize servo setup types
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.newSpecsFile, "f", "", cmdhelp.DUTUpdateFileText)

		c.Flags.StringVar(&c.hostname, "name", "", "hostname of the DUT.")
		c.Flags.StringVar(&c.machine, "asset", "", "asset tag of the DUT.")
		c.Flags.StringVar(&c.servo, "servo", "", "servo hostname and port as hostname:port. Clearing this field will delete the servo in DUT. "+cmdhelp.ClearFieldHelpText)
		c.Flags.StringVar(&c.servoSerial, "servo-serial", "", "serial number for the servo.")
		c.Flags.StringVar(&c.servoSetupType, "servo-setup", "", "servo setup type. Allowed values are "+cmdhelp.ServoSetupTypeAllowedValuesString()+".")
		c.Flags.Int64Var(&c.deployTaskTimeout, "deploy-timeout", swarming.DeployTaskExecutionTimeout, "execution timeout for deploy task in seconds.")
		c.Flags.BoolVar(&c.ignoreUFS, "ignore-ufs", false, "skip updating UFS. Starts a redeploy task.")
		c.Flags.BoolVar(&c.ignoreDeploy, "ignore-deploy", false, "skip redeploy task creation.")
		c.Flags.Var(utils.CSVString(&c.deployTags), "deploy-tags", "comma seperated tags for deployment task.")
		c.Flags.BoolVar(&c.deployDownloadImage, "deploy-download-image", false, "download image and stage usb.")
		c.Flags.BoolVar(&c.deployInstallFirmware, "deploy-install-fw", false, "install firmware.")
		c.Flags.BoolVar(&c.deployInstallOS, "deploy-install-os", false, "install os image.")
		c.Flags.BoolVar(&c.deployUpdateLabels, "deploy-update-labels", false, "update labels during deployment.")
		return c
	},
}

type updateDUT struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	newSpecsFile   string
	hostname       string
	machine        string
	servo          string
	servoSerial    string
	servoSetupType string

	ignoreDeploy          bool
	ignoreUFS             bool
	deployTaskTimeout     int64
	deployActions         []string
	deployTags            []string
	deployDownloadImage   bool
	deployInstallOS       bool
	deployInstallFirmware bool
	deployUpdateLabels    bool
}

func (c *updateDUT) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *updateDUT) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ns, err := c.envFlags.Namespace()
	if err != nil {
		return err
	}
	ctx = utils.SetupContext(ctx, ns)
	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}

	e := c.envFlags.Env()
	if c.commonFlags.Verbose() {
		if !c.ignoreUFS {
			fmt.Printf("Using UFS service %s \n", e.UnifiedFleetService)
		}
		fmt.Printf("Using swarming service %s \n", e.SwarmingService)
	}

	req, err := c.parseArgs()
	if err != nil {
		return err
	}

	// Update the UFS database if enabled.
	if !c.ignoreUFS {
		ic := ufsAPI.NewFleetPRPCClient(&prpc.Client{
			C:       hc,
			Host:    e.UnifiedFleetService,
			Options: site.DefaultPRPCOptions,
		})

		if err := c.updateDUTToUFS(ctx, ic, req); err != nil {
			return err
		}
	}

	// Swarm a deploy task if enabled.
	if !c.ignoreDeploy {
		c.updateDeployActions()
		tc, err := swarming.NewTaskCreator(ctx, &c.authFlags, e.SwarmingService)
		if err != nil {
			return err
		}
		tc.LogdogService = e.LogdogService
		tc.SwarmingServiceAccount = e.SwarmingServiceAccount
		// Start a swarming deploy task for the DUT.
		if err := c.deployDUTToSwarming(ctx, tc, req.GetMachineLSE()); err != nil {
			return err
		}
	}

	return nil
}

// validateArgs validates the set of inputs to updateDUT.
func (c updateDUT) validateArgs() error {
	if c.ignoreUFS && c.ignoreDeploy {
		return cmdlib.NewQuietUsageError(c.Flags, "Nothing to do. disable ignore-ufs or ignore-delpoy")
	}
	if c.newSpecsFile == "" && c.hostname == "" {
		return cmdlib.NewQuietUsageError(c.Flags, "Need hostname to create a DUT")
	}
	if !c.ignoreUFS && c.newSpecsFile != "" {
		// Cannot accept cmdline inputs for DUT when csv/json mode is specifed.
		if c.hostname != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe MCSV/JSON mode is specified. '-name' cannot be specified at the same time.")
		}
		if c.machine != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe MCSV/JSON mode is specified. '-asset' cannot be specified at the same time.")
		}
		if c.servo != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe MCSV/JSON mode is specified. '-servo' cannot be specified at the same time.")
		}
		if c.servoSerial != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe MCSV/JSON mode is specified. '-servo-serial' cannot be specified at the same time.")
		}
		if c.servoSetupType != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe MCSV/JSON mode is specified. '-servo-setup' cannot be specified at the same time.")
		}
		if c.servoSerial != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe MCSV/JSON mode is specified. '-servo-serial' cannot be specified at the same time.")
		}
	}
	if c.ignoreDeploy {
		// Cannot accept deploy task related inputs if it's skipped.
		if c.deployTaskTimeout != swarming.DeployTaskExecutionTimeout {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nIgnoring deploy task creation. '-deploy-timeout' cannot be specified at the same time.")
		}
		if c.deployDownloadImage {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nIgnoring deploy task creation. '-deploy-download-image' cannot be specified at the same time.")
		}
		if c.deployInstallOS {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nIgnoring deploy task creation. '-deploy-download-image' cannot be specified at the same time.")
		}
		if c.deployInstallFirmware {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nIgnoring deploy task creation. '-deploy-download-image' cannot be specified at the same time.")
		}
		if c.servo != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nIgnoring deploy task creation. '-servo' cannot be specified at the same time.")
		}
		if c.servoSerial != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nIgnoring deploy task creation. '-servo-serial' cannot be specified at the same time.")
		}
		if c.servoSetupType != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nIgnoring deploy task creation. '-servo-setup' cannot be specified at the same time.")
		}
	}
	if c.ignoreUFS {
		// Cannot accept cmdline inputs for updating DUT if we are skipping the update.
		if c.machine != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nSkipping write to UFS. '-asset' cannot be specified at the same time.")
		}
		if c.servo != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nSkipping write to UFS. '-servo' cannot be specified at the same time.")
		}
		if c.servoSerial != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nSkipping write to UFS. '-servo-serial' cannot be specified at the same time.")
		}
		if c.servoSetupType != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nSkipping write to UFS. '-servo-setup' cannot be specified at the same time.")
		}
		if c.servoSerial != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nSkipping write to UFS. '-servo-serial' cannot be specified at the same time.")
		}
	}
	if !c.ignoreUFS && c.newSpecsFile == "" {
		// Check if servo input is valid
		if c.servo != "" && c.servo != utils.ClearFieldValue {
			_, _, err := parseServoHostnamePort(c.servo)
			if err != nil {
				return err
			}
		}
		// Check if servo type is valid.
		if _, ok := lab.ServoSetupType_value[appendServoSetupPrefix(c.servoSetupType)]; c.servoSetupType != "" && !ok {
			return cmdlib.NewQuietUsageError(c.Flags, "Invalid value for servo setup type. Valid values are "+cmdhelp.ServoSetupTypeAllowedValuesString())
		}
	}
	return nil
}

// validateRequest checks if the req is valid based on the cmdline input.
func (c *updateDUT) validateRequest(ctx context.Context, ic ufsAPI.FleetClient, req *ufsAPI.UpdateMachineLSERequest) error {
	lse := req.MachineLSE
	mask := req.UpdateMask
	if lse == nil {
		return fmt.Errorf("Internal Error. Invalid UpdateMachineLSERequest")
	}
	if lse.Name == "" {
		return fmt.Errorf("Invalid update. Missing DUT name")
	}
	// Check if the mask update is valid.
	if mask != nil && len(mask.Paths) != 0 && c.ignoreDeploy {
		if containsAnyStrings(mask.Paths, "machines") {
			return fmt.Errorf("Invalid update. Cannot ignore deploy when updating asset")
		}
		// Cannot ignore deploy while doing a servo update.
		if containsAnyStrings(mask.Paths, servoPortPath, servoSetupPath, servoSerialPath, servoHostPath) {
			return fmt.Errorf("Invalid update. Cannot ignore deploy when updating servo")
		}
	}
	return nil
}

// containsAnyStrings returns true if any of the inputs are included in the list.
func containsAnyStrings(list []string, inputs ...string) bool {
	for _, a := range list {
		for _, b := range inputs {
			if b == a {
				return true
			}
		}
	}
	return false
}

// parseArgs reads input from the cmd line parameters and generates update dut request.
func (c *updateDUT) parseArgs() (*ufsAPI.UpdateMachineLSERequest, error) {
	if c.newSpecsFile != "" {
		if utils.IsCSVFile(c.newSpecsFile) {
			return nil, fmt.Errorf("Not implemented yet")
		}
		if c.ignoreDeploy {
			// Throw error if ignoring deploy task from json update.
			return nil, fmt.Errorf("JSON update. Cannot ignore deploy")
		}
		machineLse := &ufspb.MachineLSE{}
		if err := utils.ParseJSONFile(c.newSpecsFile, machineLse); err != nil {
			return nil, err
		}
		// json input updates without a mask.
		return &ufsAPI.UpdateMachineLSERequest{
			MachineLSE: machineLse,
		}, nil
	}

	lse, mask, err := c.initializeLSEAndMask()
	if err != nil {
		return nil, err
	}
	// Update delpoy actions list.
	c.updateDeployActions()
	return &ufsAPI.UpdateMachineLSERequest{
		MachineLSE: lse,
		UpdateMask: mask,
	}, nil
}

// initializeLSEAndMask reads inputs from cmd line inputs and generates LSE and corresponding mask.
func (c *updateDUT) initializeLSEAndMask() (*ufspb.MachineLSE, *field_mask.FieldMask, error) {
	var name, servoHost, servoSerial, servoSetup string
	var machines []string
	var servoPort int32
	// command line parameters
	name = c.hostname
	var err error
	if c.servo == utils.ClearFieldValue {
		// Deleting servo host is same as deleting servo
		servoHost = c.servo
	} else {
		servoHost, servoPort, err = parseServoHostnamePort(c.servo)
		if err != nil {
			return nil, nil, err
		}
	}
	servoSerial = c.servoSerial
	servoSetup = c.servoSetupType
	machines = []string{c.machine}

	// Generate lse and mask
	lse := &ufspb.MachineLSE{
		Lse: &ufspb.MachineLSE_ChromeosMachineLse{
			ChromeosMachineLse: &ufspb.ChromeOSMachineLSE{
				ChromeosLse: &ufspb.ChromeOSMachineLSE_DeviceLse{
					DeviceLse: &ufspb.ChromeOSDeviceLSE{
						Device: &ufspb.ChromeOSDeviceLSE_Dut{
							Dut: &lab.DeviceUnderTest{
								Peripherals: &lab.Peripherals{
									Servo: &lab.Servo{},
								},
							},
						},
					},
				},
			},
		},
	}
	mask := &field_mask.FieldMask{}
	lse.Name = name
	lse.Hostname = name
	lse.GetChromeosMachineLse().GetDeviceLse().GetDut().Hostname = name

	// Check if machines are being updated.
	if len(machines) > 0 && machines[0] != "" {
		lse.Machines = machines
		mask.Paths = append(mask.Paths, "machines")
	}

	// Create and assign servo and corresponding masks.
	servo, paths := generateServoWithMask(servoHost, servoSetup, servoSerial, servoPort)
	lse.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().Servo = servo
	mask.Paths = append(mask.Paths, paths...)

	// Check if nothing is being updated. Updating with an empty mask overwrites everything.
	if len(mask.Paths) == 0 || mask.Paths[0] == "" {
		return nil, nil, cmdlib.NewQuietUsageError(c.Flags, "Nothing to update")
	}
	return lse, mask, nil
}

// generateServoWithMask generates a servo object from the given inputs and corresponding mask.
func generateServoWithMask(servoHost, servoSetup, servoSerial string, servoPort int32) (*lab.Servo, []string) {
	servo := &lab.Servo{}
	paths := []string{}
	// Check and update servo port.
	if servoPort != int32(0) {
		paths = append(paths, servoPortPath)
		servo.ServoPort = servoPort
	}

	if servoSetup != "" {
		paths = append(paths, servoSetupPath)
		sst := lab.ServoSetupType(lab.ServoSetupType_value[appendServoSetupPrefix(servoSetup)])
		servo.ServoSetup = sst
	}

	if servoSerial != "" {
		paths = append(paths, servoSerialPath)
		servo.ServoSerial = servoSerial
	}

	if servoHost != "" {
		paths = append(paths, servoHostPath)
		if servoHost == utils.ClearFieldValue {
			// Delete servo for the update.
			servo = nil
		} else {
			servo.ServoHostname = servoHost
		}
	}
	return servo, paths
}

// updateDeployActions updates the deploySkipActions based on boolean skip options.
func (c *updateDUT) updateDeployActions() {
	// Append the required deploy actions
	if c.deployDownloadImage {
		c.deployActions = append(c.deployActions, "stage-usb")
	}
	if c.deployInstallOS {
		c.deployActions = append(c.deployActions, "install-test-image")
	}
	if c.deployInstallFirmware {
		c.deployActions = append(c.deployActions, "install-firmware", "verify-recovery-mode")
	}
	if c.deployInstallFirmware || c.deployInstallOS || c.deployUpdateLabels {
		c.deployActions = append(c.deployActions, "update-label")
	}
}

// updateDUTToUFS verifies the request and calls UpdateMachineLSE API with the given request.
func (c *updateDUT) updateDUTToUFS(ctx context.Context, ic ufsAPI.FleetClient, req *ufsAPI.UpdateMachineLSERequest) error {
	// Validate the update request.
	if err := c.validateRequest(ctx, ic, req); err != nil {
		return err
	}
	// Print existing LSE before update.
	if err := utils.PrintExistingHost(ctx, ic, req.MachineLSE.GetName()); err != nil {
		return err
	}
	req.MachineLSE.Name = ufsUtil.AddPrefix(ufsUtil.MachineLSECollection, req.MachineLSE.Name)

	res, err := ic.UpdateMachineLSE(ctx, req)
	if err != nil {
		return err
	}
	res.Name = ufsUtil.RemovePrefix(res.Name)
	utils.PrintProtoJSON(res, !utils.NoEmitMode(false))
	fmt.Printf("Successfully updated DUT to UFS: %s \n", res.GetName())
	return nil
}

// deployDUTToSwarming starts a re-deploy task for the given DUT.
func (c *updateDUT) deployDUTToSwarming(ctx context.Context, tc *swarming.TaskCreator, lse *ufspb.MachineLSE) error {
	var hostname, machine string
	// Using hostname because name has resource prefix
	hostname = lse.GetHostname()
	machines := lse.GetMachines()
	if len(machines) > 0 {
		machine = machines[0]
	}
	task, err := tc.DeployDut(ctx, hostname, machine, defaultSwarmingPool, c.deployTaskTimeout, c.deployActions, c.deployTags, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Triggered Deploy task for DUT %s. Follow the deploy job at %s\n", hostname, task.TaskURL)

	return nil
}
