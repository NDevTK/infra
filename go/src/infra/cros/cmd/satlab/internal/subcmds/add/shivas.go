// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package add

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/utils"
	"infra/libs/swarming"

	"infra/cros/cmd/satlab/internal/common"
	"infra/cros/cmd/satlab/internal/site"
)

// ShivasAddDUT contains all the commands for "satlab add dut" inherited from shivas.
//
// Keep this up to date with infra/cmd/shivas/ufs/subcmds/dut/add_dut.go
type shivasAddDUT struct {
	subcommands.CommandRunBase

	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	newSpecsFile   string
	hostname       string
	asset          string
	servo          string
	servoSerial    string
	servoSetupType string
	licenseTypes   []string
	licenseIds     []string
	pools          []string
	rpm            string
	rpmOutlet      string

	ignoreUFS                 bool
	deployTaskTimeout         int64
	deployActions             []string
	deployTags                []string
	deploySkipDownloadImage   bool
	deploySkipInstallOS       bool
	deploySkipInstallFirmware bool
	deploymentTicket          string
	tags                      []string
	state                     string
	description               string

	// Asset location fields
	zone string
	rack string

	// ACS DUT fields
	chameleons        []string
	cameras           []string
	antennaConnection string
	router            string
	cables            []string
	facing            string
	light             string
	carrier           string
	audioBoard        bool
	audioBox          bool
	atrus             bool
	wifiCell          bool
	touchMimo         bool
	cameraBox         bool
	chaos             bool
	audioCable        bool
	smartUSBHub       bool

	// Machine specific fields
	model string
	board string
}

// DefaultDeployTaskActions are the default actoins run at deploy time.
// TODO(gregorynisbet): this about which actions make sense for satlab.
var defaultDeployTaskActions = []string{"servo-verification", "update-label", "verify-recovery-mode", "run-pre-deploy-verification"}

// MakeDefaultShivasCommand makes the default value for the shivas portion of the
// "satlab add dut" flags.
//
// Keep this up to date with infra/cmd/shivas/ufs/subcmds/dut/add_dut.go
//
func makeDefaultShivasCommand() *addDUT {
	c := &addDUT{}
	c.pools = []string{}
	c.chameleons = []string{}
	c.cameras = []string{}
	c.cables = []string{}
	// TODO(gregorynisbet): Add more info here.
	c.deployTags = []string{"satlab"}
	// TODO(gregorynisbet): Consider skipping actions for satlab by default.
	c.deployActions = defaultDeployTaskActions
	return c
}

// Register flags inherited from shivas in place in the add DUT command.
// Keep this up to date with infra/cmd/shivas/ufs/subcmds/dut/add_dut.go
func registerShivasFlags(c *addDUT) {
	c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
	c.envFlags.Register(&c.Flags)
	c.commonFlags.Register(&c.Flags)

	c.Flags.StringVar(&c.newSpecsFile, "f", "", cmdhelp.DUTRegistrationFileText)

	// Asset location fields
	c.Flags.StringVar(&c.zone, "zone", common.DefaultZone, "Zone that the asset is in. "+cmdhelp.ZoneFilterHelpText)
	c.Flags.StringVar(&c.rack, "rack", "", "Rack that the asset is in.")

	// DUT/MachineLSE common fields
	c.Flags.StringVar(&c.hostname, "name", "", "hostname of the DUT.")
	c.Flags.StringVar(&c.asset, "asset", "", "asset tag of the machine.")
	c.Flags.StringVar(&c.servo, "servo", "", "servo hostname and port as hostname:port. (port is assigned by UFS if missing)")
	c.Flags.StringVar(&c.servoSerial, "servo-serial", "", "serial number for the servo. Can skip for Servo V3.")
	c.Flags.StringVar(&c.servoSetupType, "servo-setup", "", "servo setup type. Allowed values are "+cmdhelp.ServoSetupTypeAllowedValuesString()+", UFS assigns REGULAR if unassigned.")
	c.Flags.Var(utils.CSVString(&c.pools), "pools", "comma separated pools assigned to the DUT. 'DUT_POOL_QUOTA' is used if nothing is specified")
	c.Flags.Var(utils.CSVString(&c.licenseTypes), "licensetype", cmdhelp.LicenseTypeHelpText)
	c.Flags.Var(utils.CSVString(&c.licenseIds), "licenseid", "the name of the license type. Can specify multiple comma separated values.")
	c.Flags.StringVar(&c.rpm, "rpm", "", "rpm assigned to the DUT.")
	c.Flags.StringVar(&c.rpmOutlet, "rpm-outlet", "", "rpm outlet used for the DUT.")
	c.Flags.Int64Var(&c.deployTaskTimeout, "deploy-timeout", swarming.DeployTaskExecutionTimeout, "execution timeout for deploy task in seconds.")
	c.Flags.BoolVar(&c.ignoreUFS, "ignore-ufs", false, "skip updating UFS create a deploy task.")
	c.Flags.Var(utils.CSVString(&c.deployTags), "deploy-tags", "comma separated tags for deployment task.")
	c.Flags.BoolVar(&c.deploySkipDownloadImage, "deploy-skip-download-image", false, "skips downloading image and staging usb")
	c.Flags.BoolVar(&c.deploySkipInstallFirmware, "deploy-skip-install-fw", false, "skips installing firmware")
	c.Flags.BoolVar(&c.deploySkipInstallOS, "deploy-skip-install-os", false, "skips installing os image")
	c.Flags.StringVar(&c.deploymentTicket, "ticket", "", "the deployment ticket for this machine.")
	c.Flags.Var(utils.CSVString(&c.tags), "tags", "comma separated tags.")
	c.Flags.StringVar(&c.state, "state", "", cmdhelp.StateHelp)
	c.Flags.StringVar(&c.description, "desc", "", "description for the machine.")

	// ACS DUT fields
	c.Flags.Var(utils.CSVString(&c.chameleons), "chameleons", cmdhelp.ChameleonTypeHelpText)
	c.Flags.Var(utils.CSVString(&c.cameras), "cameras", cmdhelp.CameraTypeHelpText)
	c.Flags.Var(utils.CSVString(&c.cables), "cables", cmdhelp.CableTypeHelpText)
	c.Flags.StringVar(&c.antennaConnection, "antennaconnection", "", cmdhelp.AntennaConnectionHelpText)
	c.Flags.StringVar(&c.router, "router", "", cmdhelp.RouterHelpText)
	c.Flags.StringVar(&c.facing, "facing", "", cmdhelp.FacingHelpText)
	c.Flags.StringVar(&c.light, "light", "", cmdhelp.LightHelpText)
	c.Flags.StringVar(&c.carrier, "carrier", "", "name of the carrier.")
	c.Flags.BoolVar(&c.audioBoard, "audioboard", false, "adding this flag will specify if audioboard is present")
	c.Flags.BoolVar(&c.audioBox, "audiobox", false, "adding this flag will specify if audiobox is present")
	c.Flags.BoolVar(&c.atrus, "atrus", false, "adding this flag will specify if atrus is present")
	c.Flags.BoolVar(&c.wifiCell, "wificell", false, "adding this flag will specify if wificell is present")
	c.Flags.BoolVar(&c.touchMimo, "touchmimo", false, "adding this flag will specify if touchmimo is present")
	c.Flags.BoolVar(&c.cameraBox, "camerabox", false, "adding this flag will specify if camerabox is present")
	c.Flags.BoolVar(&c.chaos, "chaos", false, "adding this flag will specify if chaos is present")
	c.Flags.BoolVar(&c.audioCable, "audiocable", false, "adding this flag will specify if audiocable is present")
	c.Flags.BoolVar(&c.smartUSBHub, "smartusbhub", false, "adding this flag will specify if smartusbhub is present")

	// Machine fields
	// crbug.com/1188488 showed us that it might be wise to add model/board during deployment if required.
	c.Flags.StringVar(&c.model, "model", "", "model of the DUT undergoing deployment. If not given, HaRT data is used. Fails if model is not known for the DUT")
	c.Flags.StringVar(&c.board, "board", "", "board of the DUT undergoing deployment. If not given, HaRT data is used. Fails if board is not known for the DUT")
}
