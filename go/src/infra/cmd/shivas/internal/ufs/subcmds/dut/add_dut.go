// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/flag"
	"go.chromium.org/luci/grpc/prpc"
	"google.golang.org/genproto/protobuf/field_mask"

	"infra/cmd/shivas/cmdhelp"
	peripheralsCmd "infra/cmd/shivas/internal/ufs/subcmds/peripherals"
	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	"infra/libs/skylab/buildbucket"
	"infra/libs/skylab/common/heuristics"
	swarming "infra/libs/swarming"
	ufspb "infra/unifiedfleet/api/v1/models"
	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// Regexp to enforce the input format
var servoHostPortRegexp = regexp.MustCompile(`^[a-zA-Z0-9\-\.]+:[0-9]+$`)

// TODO(anushruth): Find a better place to put these tags.
var shivasTags = []string{"shivas:" + site.VersionNumber, "triggered_using:shivas"}

// defaultPools contains the list of pools used by default.
var defaultPools = []string{"DUT_POOL_QUOTA"}

// defaultSwarmingPool is the swarming pool used for all DUTs.
var defaultSwarmingPool = "ChromeOSSkylab"

// AddDUTCmd adds a MachineLSE to the database. And starts a swarming job to deploy.
var AddDUTCmd = &subcommands.Command{
	UsageLine: "dut [options ...]",
	ShortDesc: "Deploy a DUT",
	LongDesc:  cmdhelp.AddDUTLongDesc,
	CommandRun: func() subcommands.CommandRun {
		c := &addDUT{
			pools:      []string{},
			chameleons: []string{},
			cameras:    []string{},
			cables:     []string{},
			deployTags: shivasTags,
		}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.newSpecsFile, "f", "", cmdhelp.DUTRegistrationFileText)

		// Asset location fields
		c.Flags.StringVar(&c.zone, "zone", "", "Zone that the asset is in. "+cmdhelp.ZoneFilterHelpText)
		c.Flags.StringVar(&c.rack, "rack", "", "Rack that the asset is in.")

		// DUT/MachineLSE common fields
		c.Flags.StringVar(&c.hostname, "name", "", "hostname of the DUT.")
		c.Flags.StringVar(&c.asset, "asset", "", "asset tag of the machine.")
		c.Flags.StringVar(&c.servo, "servo", "", "servo hostname and port as hostname:port. (port is assigned by UFS if missing)")
		c.Flags.StringVar(&c.servoSerial, "servo-serial", "", "serial number for the servo. Can skip for Servo V3.")
		c.Flags.StringVar(&c.servoSetupType, "servo-setup", "", "servo setup type. Allowed values are "+cmdhelp.ServoSetupTypeAllowedValuesString()+", UFS assigns REGULAR if unassigned.")
		c.Flags.StringVar(&c.servoDockerContainerName, "servod-docker", "", "servod docker container name. Required if serovd is running on docker")
		c.Flags.Var(utils.CSVString(&c.pools), "pools", "comma separated pools assigned to the DUT. 'DUT_POOL_QUOTA' is used if nothing is specified")
		c.Flags.Var(utils.CSVString(&c.licenseTypes), "licensetype", cmdhelp.LicenseTypeHelpText)
		c.Flags.Var(utils.CSVString(&c.licenseIds), "licenseid", "the name of the license type. Can specify multiple comma separated values.")
		c.Flags.StringVar(&c.rpm, "rpm", "", "rpm assigned to the DUT.")
		c.Flags.StringVar(&c.rpmOutlet, "rpm-outlet", "", "rpm outlet used for the DUT.")
		c.Flags.BoolVar(&c.ignoreUFS, "ignore-ufs", false, "skip updating UFS create a deploy task.")
		c.Flags.Var(utils.CSVString(&c.deployTags), "deploy-tags", "comma separated tags for deployment task.")
		c.Flags.StringVar(&c.deploymentTicket, "ticket", "", "the deployment ticket for this machine.")
		c.Flags.Var(flag.StringSlice(&c.tags), "tag", "Name(s) of tag(s). Can be specified multiple times.")
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
		c.Flags.BoolVar(&c.appcap, "appcap", false, "adding this flag will specify if lab wifi ap/pcap is present")
		c.Flags.BoolVar(&c.touchMimo, "touchmimo", false, "adding this flag will specify if touchmimo is present")
		c.Flags.BoolVar(&c.cameraBox, "camerabox", false, "adding this flag will specify if camerabox is present")
		c.Flags.BoolVar(&c.chaos, "chaos", false, "adding this flag will specify if chaos is present")
		c.Flags.BoolVar(&c.audioCable, "audiocable", false, "adding this flag will specify if audiocable is present")
		c.Flags.BoolVar(&c.smartUSBHub, "smartusbhub", false, "adding this flag will specify if smartusbhub is present")

		// Machine fields
		// crbug.com/1188488 showed us that it might be wise to add model/board during deployment if required.
		c.Flags.StringVar(&c.model, "model", "", "model of the DUT undergoing deployment. If not given, HaRT data is used. Fails if model is not known for the DUT")
		c.Flags.StringVar(&c.board, "board", "", "board of the DUT undergoing deployment. If not given, HaRT data is used. Fails if board is not known for the DUT")

		// Multi-peripherals
		c.Flags.UintVar(&c.bluetoothPeersCount, "btpn", 0, "number of Bluetooth peers connected")

		// Scheduling
		c.Flags.BoolVar(&c.latestVersion, "latest", false, "Use latest version of CIPD when scheduling. By default use prod.")
		return c
	},
}

type addDUT struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	newSpecsFile             string
	hostname                 string
	asset                    string
	servo                    string
	servoSerial              string
	servoSetupType           string
	servoDockerContainerName string
	licenseTypes             []string
	licenseIds               []string
	pools                    []string
	rpm                      string
	rpmOutlet                string

	ignoreUFS        bool
	deployTags       []string
	deploymentTicket string
	tags             []string
	state            string
	description      string

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
	appcap            bool
	touchMimo         bool
	cameraBox         bool
	chaos             bool
	audioCable        bool
	smartUSBHub       bool

	// Machine specific fields
	model string
	board string

	// Multi-peripherals
	bluetoothPeersCount uint

	// Scheduling
	latestVersion bool
}

var mcsvFields = []string{
	"name",
	"asset",
	"model",
	"board",
	"servo_host",
	"servo_port",
	"servo_serial",
	"servo_setup",
	"rpm_host",
	"rpm_outlet",
	"pools",
}

// dutDeployUFSParams contains all the data that are needed for deployment of a single DUT
// Asset and its update paths are required here to update location, model and board for the DUT
// See: crbug.com/1188488 for why model and board need to be updated.
type dutDeployUFSParams struct {
	DUT   *ufspb.MachineLSE // MachineLSE of the DUT to be updated
	Asset *ufspb.Asset      // Asset underlying the DUT being updated
	Paths []string          // Update paths for the Asset being updated
}

func (c *addDUT) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *addDUT) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ns, err := c.getNamespace()
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

	tc, err := swarming.NewTaskCreator(ctx, &c.authFlags, e.SwarmingService)
	if err != nil {
		return err
	}
	tc.LogdogService = e.LogdogService
	tc.SwarmingServiceAccount = e.SwarmingServiceAccount

	dutParams, err := c.parseArgs()
	if err != nil {
		return err
	}
	bc, err := buildbucket.NewClient(ctx, hc, site.DefaultPRPCOptions, "chromeos", "labpack", "labpack")
	if err != nil {
		return err
	}
	sessionTag := fmt.Sprintf("admin-session:%s", uuid.New().String())

	// Created client to update UFS when required.
	var ic ufsAPI.FleetClient
	if !c.ignoreUFS {
		ic = ufsAPI.NewFleetPRPCClient(&prpc.Client{
			C:       hc,
			Host:    e.UnifiedFleetService,
			Options: site.DefaultPRPCOptions,
		})
	}
	for _, param := range dutParams {
		if !c.ignoreUFS {
			// Update the UFS database if enabled.
			if len(param.DUT.GetMachines()) == 0 {
				fmt.Printf("Failed to add DUT %s to UFS. It is not linked to any Asset(Machine).\n", param.DUT.GetName())
				continue
			}
			if err := c.addDutToUFS(ctx, ic, param); err != nil {
				fmt.Printf("Failed to add DUT %s to UFS. Skipping deployment. %s", param.DUT.GetName(), err.Error())
				// skip deployment
				continue
			}
		}
		utils.ScheduleDeployTask(ctx, bc, e, param.DUT.GetName(), sessionTag, c.latestVersion)
	}
	if len(dutParams) > 1 {
		fmt.Fprintf(a.GetOut(), "\nBatch tasks URL: %s\n\n", utils.TasksBatchLink(e.SwarmingService, sessionTag))
	}
	return nil
}

// getNamespace returns the namespace used to call UFS with appropriate
// validation and default behavior. It is primarily separated from the main
// function for testing purposes
func (c *addDUT) getNamespace() (string, error) {
	return c.envFlags.Namespace(site.OSLikeNamespaces, ufsUtil.OSNamespace)
}

func (c addDUT) validateArgs() error {
	if !c.ignoreUFS && c.newSpecsFile == "" {
		if c.asset == "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Need asset ID to create a DUT")
		}
		if c.servo != "" {
			// If the servo is not servo V3. Then servo serial is needed
			host, _, err := parseServoHostnamePort(c.servo)
			if err != nil {
				return err
			}
			if !ufsUtil.ServoV3HostnameRegex.MatchString(host) && c.servoSerial == "" {
				return cmdlib.NewQuietUsageError(c.Flags, "Cannot skip servo serial. Not a servo V3 device.")
			}
		} else {
			if !(ufsUtil.IsChromiumLegacyHost(c.hostname) || ufsUtil.IsChromeLegacyHost(c.hostname)) && (c.servoSerial != "" || c.servoSetupType != "" || c.servoDockerContainerName != "") {
				return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nProvided extra servo details when servo hostname is not provided.")
			}
		}
		if c.servoSetupType != "" {
			if _, ok := chromeosLab.ServoSetupType_value[appendServoSetupPrefix(c.servoSetupType)]; !ok {
				return cmdlib.NewQuietUsageError(c.Flags, "Invalid servo setup %s", c.servoSetupType)
			}
		}
		if (c.rpm != "" && c.rpmOutlet == "") || (c.rpm == "" && c.rpmOutlet != "") {
			return cmdlib.NewQuietUsageError(c.Flags, "Need both rpm and its outlet. %s:%s is invalid", c.rpm, c.rpmOutlet)
		}
		if c.zone != "" && !ufsUtil.IsUFSZone(ufsUtil.RemoveZonePrefix(c.zone)) {
			return cmdlib.NewQuietUsageError(c.Flags, "Invalid zone %s", c.zone)
		}
		if len(c.licenseTypes) != len(c.licenseIds) {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nNumber of -licensetype(%s) and -licenseid(%s) must be same.", c.licenseTypes, c.licenseIds)
		}
		for _, cp := range c.licenseTypes {
			if !ufsUtil.IsLicenseType(cp) {
				return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n%s is not a valid license type name, please check help info for '-licensetype'.", cp)
			}
		}
		for _, cp := range c.chameleons {
			if !ufsUtil.IsChameleonType(cp) {
				return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n%s is not a valid chameleon type name, please check help info for '-chameleons'.", cp)
			}
		}
		for _, cp := range c.cameras {
			if !ufsUtil.IsCameraType(cp) {
				return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n%s is not a valid camera type name, please check help info for '-cameras'.", cp)
			}
		}
		for _, cp := range c.cables {
			if !ufsUtil.IsCableType(cp) {
				return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n%s is not a valid cable type name, please check help info for '-cables'.", cp)
			}
		}
		if c.antennaConnection != "" && !ufsUtil.IsAntennaConnection(c.antennaConnection) {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n%s is not a valid antenna connection name, please check help info for '-antennaconnection'.", c.antennaConnection)
		}
		if c.router != "" && !ufsUtil.IsRouter(c.router) {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n%s is not a valid router name, please check help info for '-router'.", c.router)
		}
		if c.facing != "" && !ufsUtil.IsFacing(c.facing) {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n%s is not a valid facing name, please check help info for '-facing'.", c.facing)
		}
		if c.light != "" && !ufsUtil.IsLight(c.light) {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n%s is not a valid light name, please check help info for '-light'.", c.light)
		}
	}
	if c.newSpecsFile == "" && c.hostname == "" {
		return cmdlib.NewQuietUsageError(c.Flags, "Need hostname to create a DUT")
	}

	if c.bluetoothPeersCount > 4 {
		return cmdlib.NewQuietUsageError(c.Flags, "Too many Bluetooth peers specified: %d", c.bluetoothPeersCount)
	}
	return nil
}

func (c *addDUT) parseArgs() ([]*dutDeployUFSParams, error) {
	if c.newSpecsFile != "" {
		if utils.IsCSVFile(c.newSpecsFile) {
			return c.parseMCSV()
		}
		machinelse := &ufspb.MachineLSE{}
		if err := utils.ParseJSONFile(c.newSpecsFile, machinelse); err != nil {
			return nil, err
		}
		asset, paths := utils.GenerateAssetUpdate(machinelse.GetMachines()[0], c.model, c.board, c.zone, c.rack)
		return []*dutDeployUFSParams{{
			DUT:   machinelse,
			Asset: asset,
			Paths: paths,
		}}, nil
	}
	// command line parameters
	dutParams, err := c.initializeLSEAndAsset(nil)
	if err != nil {
		return nil, err
	}
	return []*dutDeployUFSParams{dutParams}, nil
}

// parseMCSV parses the MCSV file and returns MachineLSEs
func (c *addDUT) parseMCSV() ([]*dutDeployUFSParams, error) {
	records, err := utils.ParseMCSVFile(c.newSpecsFile)
	if err != nil {
		return nil, err
	}
	var dutParams []*dutDeployUFSParams
	for i, rec := range records {
		// if i is 1, determine whether this is a header
		if i == 0 && heuristics.LooksLikeHeader(rec) {
			if err := utils.ValidateSameStringArray(mcsvFields, rec); err != nil {
				return nil, err
			}
			continue
		}
		recMap := make(map[string]string)
		for j, title := range mcsvFields {
			recMap[title] = rec[j]
		}
		p, err := c.initializeLSEAndAsset(recMap)
		if err != nil {
			fmt.Printf("Error [%s:%v]: %v. Skipping add on this line\n", c.newSpecsFile, i+1, err.Error())
		} else {
			dutParams = append(dutParams, p)
		}
	}
	return dutParams, nil
}

func (c *addDUT) addDutToUFS(ctx context.Context, ic ufsAPI.FleetClient, param *dutDeployUFSParams) error {
	// Attempt to update the changes to asset first.
	if err := c.updateAssetToUFS(ctx, ic, param.Asset, param.Paths); err != nil {
		return err
	}
	if !ufsUtil.ValidateTags(param.DUT.Tags) {
		return fmt.Errorf(ufsAPI.InvalidTags)
	}
	res, err := ic.CreateMachineLSE(ctx, &ufsAPI.CreateMachineLSERequest{
		MachineLSE:   param.DUT,
		MachineLSEId: param.DUT.GetName(),
	})
	if err != nil {
		fmt.Printf("Failed to add DUT %s to UFS. UFS add failed %s\n", param.DUT.GetName(), err)
		return err
	}
	res.Name = ufsUtil.RemovePrefix(res.Name)
	utils.PrintProtoJSON(res, !utils.NoEmitMode(false))
	fmt.Printf("Successfully added DUT to UFS: %s \n", res.GetName())
	return nil
}

func (c *addDUT) initializeLSEAndAsset(recMap map[string]string) (*dutDeployUFSParams, error) {
	lse := &ufspb.MachineLSE{
		Lse: &ufspb.MachineLSE_ChromeosMachineLse{
			ChromeosMachineLse: &ufspb.ChromeOSMachineLSE{
				ChromeosLse: &ufspb.ChromeOSMachineLSE_DeviceLse{
					DeviceLse: &ufspb.ChromeOSDeviceLSE{
						Device: &ufspb.ChromeOSDeviceLSE_Dut{
							Dut: &chromeosLab.DeviceUnderTest{
								Peripherals: &chromeosLab.Peripherals{
									Chameleon:     &chromeosLab.Chameleon{},
									Servo:         &chromeosLab.Servo{},
									Rpm:           &chromeosLab.OSRPM{},
									Audio:         &chromeosLab.Audio{},
									Wifi:          &chromeosLab.Wifi{},
									Touch:         &chromeosLab.Touch{},
									CameraboxInfo: &chromeosLab.Camerabox{},
								},
							},
						},
					},
				},
			},
		},
	}
	var name, servoHost, servoSerial, rpmHost, rpmOutlet, model, board string
	var pools, machines []string
	var servoPort int32
	var servoSetup chromeosLab.ServoSetupType
	resourceState := ufsUtil.ToUFSState(c.state)
	if recMap != nil {
		// CSV map
		name = recMap["name"]
		servoHost = recMap["servo_host"]
		if recMap["servo_port"] != "" {
			port, err := strconv.ParseInt(recMap["servo_port"], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("failed to parse servo port %s. %s", recMap["servo_port"], err)
			}
			servoPort = int32(port)
		}
		servoSerial = recMap["servo_serial"]
		// Check if the host is servo V3. Need servo serial otherwise.
		if !(ufsUtil.IsChromiumLegacyHost(name) || ufsUtil.IsChromeLegacyHost(name)) && !ufsUtil.ServoV3HostnameRegex.MatchString(servoHost) && servoSerial == "" {
			return nil, fmt.Errorf("Not a servo V3 host[%s]. Need servo serial", servoHost)
		}
		sst, ok := chromeosLab.ServoSetupType_value[appendServoSetupPrefix(recMap["servo_setup"])]
		if !ok && recMap["servo_setup"] != "" {
			return nil, fmt.Errorf("Invalid servo setup %s. Valid types are %s", recMap["servo_setup"], cmdhelp.ServoSetupTypeAllowedValuesString())
		}
		servoSetup = chromeosLab.ServoSetupType(sst) // Default value is REGULAR(0).
		rpmHost = recMap["rpm_host"]
		rpmOutlet = recMap["rpm_outlet"]
		machines = []string{recMap["asset"]}
		pools = strings.Fields(recMap["pools"])
		model = recMap["model"]
		board = recMap["board"]
	} else {
		// command line parameters
		name = c.hostname
		var err error
		if c.servo != "" {
			servoHost, servoPort, err = parseServoHostnamePort(c.servo)
			if err != nil {
				return nil, err
			}
			servoSerial = c.servoSerial
			servoSetup = chromeosLab.ServoSetupType(chromeosLab.ServoSetupType_value[appendServoSetupPrefix(c.servoSetupType)])
		}
		rpmHost = c.rpm
		rpmOutlet = c.rpmOutlet
		machines = []string{c.asset}
		pools = c.pools
		model = c.model
		board = c.board
	}
	lse.Name = name
	lse.Hostname = name
	lse.GetChromeosMachineLse().GetDeviceLse().GetDut().Hostname = name
	lse.Machines = machines

	// Use the input params if available for all the options.
	lse.Description = c.description
	lse.DeploymentTicket = c.deploymentTicket
	lse.Tags = c.tags
	lse.ResourceState = resourceState

	licenses := make([]*chromeosLab.License, 0, len(c.licenseTypes))
	for i := range c.licenseTypes {
		licenses = append(licenses, &chromeosLab.License{
			Type:       ufsUtil.ToLicenseType(c.licenseTypes[i]),
			Identifier: c.licenseIds[i],
		})
	}
	lse.GetChromeosMachineLse().GetDeviceLse().GetDut().Licenses = licenses
	peripherals := lse.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals()
	if servoHost != "" {
		// if servo-host is not provided then do not set any servo field.
		peripherals.GetServo().ServoHostname = servoHost
		peripherals.GetServo().ServoPort = servoPort
		peripherals.GetServo().ServoSerial = servoSerial
		peripherals.GetServo().ServoSetup = servoSetup
		peripherals.GetServo().DockerContainerName = c.servoDockerContainerName
	}
	peripherals.GetRpm().PowerunitName = rpmHost
	peripherals.GetRpm().PowerunitOutlet = rpmOutlet
	if len(pools) > 0 && pools[0] != "" {
		lse.GetChromeosMachineLse().GetDeviceLse().GetDut().Pools = pools
	} else {
		lse.GetChromeosMachineLse().GetDeviceLse().GetDut().Pools = defaultPools
	}

	// ACS DUT fields
	chameleons := make([]chromeosLab.ChameleonType, 0, len(c.chameleons))
	for _, cp := range c.chameleons {
		chameleons = append(chameleons, ufsUtil.ToChameleonType(cp))
	}
	cameras := make([]*chromeosLab.Camera, 0, len(c.cameras))
	for _, cp := range c.cameras {
		camera := &chromeosLab.Camera{
			CameraType: ufsUtil.ToCameraType(cp),
		}
		cameras = append(cameras, camera)
	}
	cables := make([]*chromeosLab.Cable, 0, len(c.cables))
	for _, cp := range c.cables {
		cable := &chromeosLab.Cable{
			Type: ufsUtil.ToCableType(cp),
		}
		cables = append(cables, cable)
	}

	var bluetoothPeers []*chromeosLab.BluetoothPeer
	if c.bluetoothPeersCount > 0 {
		for i := uint(1); i <= c.bluetoothPeersCount; i++ {
			btpHost := fmt.Sprintf("%s-btpeer%d", c.hostname, i)
			bluetoothPeers = append(bluetoothPeers, peripheralsCmd.CreateBluetoothPeer(btpHost))
		}
	}
	peripherals.GetChameleon().ChameleonPeripherals = chameleons
	peripherals.ConnectedCamera = cameras
	peripherals.Cable = cables
	peripherals.GetWifi().AntennaConn = ufsUtil.ToAntennaConnection(c.antennaConnection)
	peripherals.GetWifi().Router = ufsUtil.ToRouter(c.router)
	peripherals.GetCameraboxInfo().Facing = ufsUtil.ToFacing(c.facing)
	peripherals.GetCameraboxInfo().Light = ufsUtil.ToLight(c.light)
	peripherals.GetChameleon().AudioBoard = c.audioBoard
	peripherals.GetAudio().AudioBox = c.audioBox
	peripherals.GetAudio().Atrus = c.atrus
	peripherals.GetAudio().AudioCable = c.audioCable
	peripherals.GetWifi().Wificell = c.wifiCell
	peripherals.GetTouch().Mimo = c.touchMimo
	peripherals.Carrier = c.carrier
	peripherals.Camerabox = c.cameraBox
	peripherals.Chaos = c.chaos
	peripherals.SmartUsbhub = c.smartUSBHub
	peripherals.BluetoothPeers = bluetoothPeers
	if c.appcap {
		peripherals.GetWifi().WifiRouters = []*chromeosLab.WifiRouter{
			{
				Hostname:    fmt.Sprintf("%s-router", name),
				BuildTarget: "gale",
				Model:       "gale",
			},
			{
				Hostname:    fmt.Sprintf("%s-pcap", name),
				BuildTarget: "gale",
				Model:       "gale",
			},
		}
	}
	// Get the updated asset and update paths
	asset, paths := utils.GenerateAssetUpdate(machines[0], model, board, c.zone, c.rack)
	return &dutDeployUFSParams{
		DUT:   lse,
		Asset: asset,
		Paths: paths,
	}, nil
}

func parseServoHostnamePort(servo string) (string, int32, error) {
	var servoHostname string
	var servoPort int32
	servoHostnamePort := strings.Split(servo, ":")
	if len(servoHostnamePort) == 2 {
		servoHostname = servoHostnamePort[0]
		port, err := strconv.ParseInt(servoHostnamePort[1], 10, 32)
		if err != nil {
			return "", int32(0), err
		}
		servoPort = int32(port)
	} else {
		servoHostname = servoHostnamePort[0]
	}
	return servoHostname, servoPort, nil
}

func appendServoSetupPrefix(servoSetup string) string {
	return fmt.Sprintf("SERVO_SETUP_%s", servoSetup)
}

// updateAssetToUFS calls UpdateAsset API in UFS with asset and partial paths
func (c *addDUT) updateAssetToUFS(ctx context.Context, ic ufsAPI.FleetClient, asset *ufspb.Asset, paths []string) error {
	if len(paths) == 0 {
		// If no update is available. Skip doing anything
		return nil
	}
	mask := &field_mask.FieldMask{
		Paths: paths,
	}
	_, err := ic.UpdateAsset(ctx, &ufsAPI.UpdateAssetRequest{
		Asset:      asset,
		UpdateMask: mask,
	})
	return err
}

// isUsingServo returns true if both servoHostName and servoSerial are defined in peripherals.servo
func isUsingServo(dutLSE *ufspb.MachineLSE) bool {
	peripherals := dutLSE.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals()
	if peripherals.GetServo().GetServoSerial() == "" && peripherals.GetServo().GetServoHostname() == "" {
		return false
	}
	return true
}
