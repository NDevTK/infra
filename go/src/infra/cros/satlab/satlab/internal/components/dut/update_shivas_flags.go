// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"strings"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/utils"
	"infra/cros/satlab/common/site"
)

// MakeUpdateShivasFlags serializes the command line arguments of updateDUT into a flagmap
// so that it can be used to call shivas directly.
func makeUpdateShivasFlags(c *updateDUT) flagmap {
	out := make(flagmap)

	if c.newSpecsFile != "" {
		// do nothing
	}
	if c.hostname != "" {
		// Do nothing. The hostname must be qualified when
		// passed to shivas.
	}
	if c.machine != "" {
		// Do nothing. The asset tag must be qualified when passed to shivas.
	}
	if c.servo != "" {
		// Do nothing.
		// The servo must be qualified when passed to shivas.
	}
	if c.servoSerial != "" {
		out["servo-serial"] = []string{c.servoSerial}
	}
	if c.servoSetupType != "" {
		out["servo-setup"] = []string{c.servoSetupType}
	}
	if c.servoDockerContainerName != "" {
		out["servod-docker"] = []string{c.servoDockerContainerName}
	}
	if len(c.pools) != 0 {
		out["pools"] = []string{strings.Join(c.pools, ",")}
	}
	if len(c.licenseTypes) != 0 {
		out["licensetype"] = []string{strings.Join(c.licenseTypes, ",")}
	}
	if c.rpm != "" {
		out["rpm"] = []string{c.rpm}
	}
	if c.rpmOutlet != "" {
		out["rpm-outlet"] = []string{c.rpmOutlet}
	}
	if len(c.deployTags) != 0 {
		out["deploy-tags"] = []string{strings.Join(c.deployTags, ",")}
	}
	if len(c.tags) != 0 {
		out["tags"] = []string{strings.Join(c.tags, ",")}
	}
	if c.description != "" {
		out["desc"] = []string{c.description}
	}
	if len(c.chameleons) != 0 {
		out["chameleons"] = []string{strings.Join(c.chameleons, ",")}
	}
	if len(c.cameras) != 0 {
		out["cameras"] = []string{strings.Join(c.cameras, ",")}
	}
	if len(c.cables) != 0 {
		out["cables"] = []string{strings.Join(c.cables, ",")}
	}
	if c.antennaConnection != "" {
		out["antennaconnection"] = []string{c.antennaConnection}
	}
	if c.router != "" {
		out["router"] = []string{c.router}
	}
	if c.facing != "" {
		out["facing"] = []string{c.facing}
	}
	if c.light != "" {
		out["light"] = []string{c.light}
	}
	if c.carrier != "" {
		out["carrier"] = []string{c.carrier}
	}
	if c.audioBoard {
		out["audioboard"] = []string{}
	}
	if c.audioBox {
		out["audiobox"] = []string{}
	}
	if c.atrus {
		out["atrus"] = []string{}
	}
	if c.wifiCell {
		out["wificell"] = []string{}
	}
	if c.touchMimo {
		out["touchmimo"] = []string{}
	}
	if c.cameraBox {
		out["camerabox"] = []string{}
	}
	if c.chaos {
		out["chaos"] = []string{}
	}
	if c.audioCable {
		out["audiocable"] = []string{}
	}
	if c.smartUSBHub {
		out["smartusbhub"] = []string{}
	}
	if c.envFlags.GetNamespace() != "" {
		// Do nothing.
	}
	return out
}

// ShivasUpdateDUT is a command that contains the arguments that "shivas update" understands.
type shivasUpdateDUT struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	// DUT specification inputs.
	newSpecsFile             string
	hostname                 string
	machine                  string
	servo                    string
	servoSerial              string
	servoSetupType           string
	servoFwChannel           string
	servoDockerContainerName string
	pools                    []string
	licenseTypes             []string
	licenseIds               []string
	rpm                      string
	rpmOutlet                string
	deploymentTicket         string
	tags                     []string
	description              string

	// Deploy task inputs.
	forceDeploy bool
	deployTags  []string

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

	// For use in determining if a flag is set
	flagInputs map[string]bool
}

// RegisterUpdateShivasFlags registers the flags inherited from shivas.
func registerUpdateShivasFlags(c *updateDUT) {
	c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
	c.envFlags.Register(&c.Flags)
	c.commonFlags.Register(&c.Flags)

	c.Flags.StringVar(&c.newSpecsFile, "f", "", cmdhelp.DUTUpdateFileText)

	c.Flags.StringVar(&c.hostname, "name", "", "hostname of the DUT.")
	c.Flags.StringVar(&c.machine, "asset", "", "asset tag of the DUT.")
	c.Flags.StringVar(&c.servo, "servo", "", "servo hostname and port as hostname:port. Clearing this field will delete the servo in DUT. "+cmdhelp.ClearFieldHelpText)
	c.Flags.StringVar(&c.servoSerial, "servo-serial", "", "serial number for the servo.")
	c.Flags.StringVar(&c.servoSetupType, "servo-setup", "", "servo setup type. Allowed values are "+cmdhelp.ServoSetupTypeAllowedValuesString()+".")
	c.Flags.StringVar(&c.servoFwChannel, "servo-fw-channel", "", "servo firmware channel. Allowed values are "+cmdhelp.ServoFwChannelAllowedValuesString()+".")
	c.Flags.StringVar(&c.servoDockerContainerName, "servod-docker", "", "servo docker container name. Required if servod is running in docker.")
	c.Flags.Var(utils.CSVString(&c.pools), "pools", "comma seperated pools. These will be appended to existing pools. "+cmdhelp.ClearFieldHelpText)
	c.Flags.Var(utils.CSVString(&c.licenseTypes), "licensetype", cmdhelp.LicenseTypeHelpText)
	c.Flags.Var(utils.CSVString(&c.licenseIds), "licenseid", "the name of the license type. Can specify multiple comma separated values. "+cmdhelp.ClearFieldHelpText)
	c.Flags.StringVar(&c.rpm, "rpm", "", "rpm assigned to the DUT. Clearing this field will delete rpm. "+cmdhelp.ClearFieldHelpText)
	c.Flags.StringVar(&c.rpmOutlet, "rpm-outlet", "", "rpm outlet used for the DUT.")
	c.Flags.StringVar(&c.deploymentTicket, "ticket", "", "the deployment ticket for this machine. "+cmdhelp.ClearFieldHelpText)
	c.Flags.Var(utils.CSVString(&c.tags), "tags", "comma separated tags. You can only append new tags or delete all of them. "+cmdhelp.ClearFieldHelpText)
	c.Flags.StringVar(&c.description, "desc", "", "description for the machine. "+cmdhelp.ClearFieldHelpText)

	c.Flags.BoolVar(&c.forceDeploy, "force-deploy", false, "forces a deploy task for all the updates.")
	c.Flags.Var(utils.CSVString(&c.deployTags), "deploy-tags", "comma seperated tags for deployment task.")

	// ACS DUT fields
	c.Flags.Var(utils.CSVString(&c.chameleons), "chameleons", cmdhelp.ChameleonTypeHelpText+". "+cmdhelp.ClearFieldHelpText)
	c.Flags.Var(utils.CSVString(&c.cameras), "cameras", cmdhelp.CameraTypeHelpText+". "+cmdhelp.ClearFieldHelpText)
	c.Flags.Var(utils.CSVString(&c.cables), "cables", cmdhelp.CableTypeHelpText+". "+cmdhelp.ClearFieldHelpText)
	c.Flags.StringVar(&c.antennaConnection, "antennaconnection", "", cmdhelp.AntennaConnectionHelpText)
	c.Flags.StringVar(&c.router, "router", "", cmdhelp.RouterHelpText)
	c.Flags.StringVar(&c.facing, "facing", "", cmdhelp.FacingHelpText)
	c.Flags.StringVar(&c.light, "light", "", cmdhelp.LightHelpText)
	c.Flags.StringVar(&c.carrier, "carrier", "", "name of the carrier."+". "+cmdhelp.ClearFieldHelpText)
	c.Flags.BoolVar(&c.audioBoard, "audioboard", false, "adding this flag will specify if audioboard is present")
	c.Flags.BoolVar(&c.audioBox, "audiobox", false, "adding this flag will specify if audiobox is present")
	c.Flags.BoolVar(&c.atrus, "atrus", false, "adding this flag will specify if atrus is present")
	c.Flags.BoolVar(&c.wifiCell, "wificell", false, "adding this flag will specify if wificell is present")
	c.Flags.BoolVar(&c.touchMimo, "touchmimo", false, "adding this flag will specify if touchmimo is present")
	c.Flags.BoolVar(&c.cameraBox, "camerabox", false, "adding this flag will specify if camerabox is present")
	c.Flags.BoolVar(&c.chaos, "chaos", false, "adding this flag will specify if chaos is present")
	c.Flags.BoolVar(&c.audioCable, "audiocable", false, "adding this flag will specify if audiocable is present")
	c.Flags.BoolVar(&c.smartUSBHub, "smartusbhub", false, "adding this flag will specify if smartusbhub is present")
}
