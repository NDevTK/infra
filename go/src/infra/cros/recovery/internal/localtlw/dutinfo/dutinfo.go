// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package dutinfo provides help function to work with DUT info.
package dutinfo

import (
	"fmt"
	"runtime/debug"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/dutstate"
	"infra/cros/recovery/tlw"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsdevice "infra/unifiedfleet/api/v1/models/chromeos/device"
	ufslab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	ufsmake "infra/unifiedfleet/api/v1/models/chromeos/manufacturing"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// ConvertDut converts USF data to local representation of Dut instance.
func ConvertDut(data *ufspb.ChromeOSDeviceData) (dut *tlw.Dut, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Reason("convert dut: %v\n%s", r, debug.Stack()).Err()
		}
	}()
	// TODO(otabek@): Add logic to read and update state file on the drones. (ProvisionedInfo)
	if data.GetLabConfig().GetChromeosMachineLse().GetDeviceLse().GetDut() != nil {
		return adaptUfsDutToTLWDut(data)
	} else if data.GetLabConfig().GetChromeosMachineLse().GetDeviceLse().GetLabstation() != nil {
		return adaptUfsLabstationToTLWDut(data)
	} else if data.GetLabConfig().GetChromeosMachineLse().GetDeviceLse().GetDevboard() != nil {
		return adaptUfsDevBoardToTLWDut(data)
	}
	return nil, errors.Reason("convert dut: unexpected case!").Err()
}

// ConvertAttachedDeviceToTlw converts USF data to local representation of Dut instance.
func ConvertAttachedDeviceToTlw(data *ufsAPI.AttachedDeviceData) (dut *tlw.Dut, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Reason("convert dut: %v\n%s", r, debug.Stack()).Err()
		}
	}()
	machine := data.GetMachine()
	machineLSE := data.GetLabConfig()
	if machine == nil || machineLSE == nil {
		return nil, errors.Reason("convert attached device to tlw: unexpected case!").Err()
	}
	// Determine type of device.
	setup := tlw.DUTSetupTypeUnspecified
	switch dt := machine.GetAttachedDevice().GetDeviceType(); dt {
	case ufspb.AttachedDeviceType_ATTACHED_DEVICE_TYPE_ANDROID_PHONE, ufspb.AttachedDeviceType_ATTACHED_DEVICE_TYPE_ANDROID_TABLET:
		setup = tlw.DUTSetupTypeAndroid
	// case ufspb.AttachedDeviceType_ATTACHED_DEVICE_TYPE_APPLE_PHONE, ufspb.AttachedDeviceType_ATTACHED_DEVICE_TYPE_APPLE_TABLET:
	// 	setup = tlw.DUTSetupTypeIOS
	default:
		panic(fmt.Sprintf("Not supported device type %q", dt.String()))
	}
	return &tlw.Dut{
		Id:   machine.GetName(),
		Name: machineLSE.GetHostname(),
		Android: &tlw.Android{
			Board:              machine.GetAttachedDevice().GetBuildTarget(),
			Model:              machine.GetAttachedDevice().GetModel(),
			SerialNumber:       machine.GetSerialNumber(),
			AssociatedHostname: machineLSE.GetAttachedDeviceLse().GetAssociatedHostname(),
		},
		SetupType:       setup,
		State:           dutstate.ConvertFromUFSState(machineLSE.GetResourceState()),
		ExtraAttributes: map[string][]string{},
		ProvisionedInfo: &tlw.ProvisionedInfo{},
		// DutStateReason not supported.
	}, nil
}

// CreateUpdateDutRequest creates request instance to update UFS.
func CreateUpdateDutRequest(dutID string, dut *tlw.Dut) (req *ufsAPI.UpdateDeviceRecoveryDataRequest, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Reason("update dut specs: %v\n%s", r, debug.Stack()).Err()
		}
	}()
	if dut.GetChromeos() != nil {
		return &ufsAPI.UpdateDeviceRecoveryDataRequest{
			DeviceId:      dutID,
			Hostname:      dut.Name,
			ResourceType:  ufsAPI.UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_CHROMEOS_DEVICE,
			ResourceState: dutstate.ConvertToUFSState(dut.State),
			DeviceRecoveryData: &ufsAPI.UpdateDeviceRecoveryDataRequest_Chromeos{
				Chromeos: &ufsAPI.ChromeOsRecoveryData{
					DutState: getUFSDutComponentStateFromSpecs(dutID, dut),
					DutData:  getUFSDutDataFromSpecs(dut),
					LabData:  getUFSLabDataFromSpecs(dut),
				},
			},
		}, nil
	} else if dut.GetAndroid() != nil {
		return &ufsAPI.UpdateDeviceRecoveryDataRequest{
			DeviceId:      dutID,
			Hostname:      dut.Name,
			ResourceType:  ufsAPI.UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_ATTACHED_DEVICE,
			ResourceState: dutstate.ConvertToUFSState(dut.State),
		}, nil
	} else if devboard := dut.GetDevBoard(); devboard != nil {
		return &ufsAPI.UpdateDeviceRecoveryDataRequest{
			DeviceId:      dutID,
			Hostname:      dut.Name,
			ResourceType:  ufsAPI.UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_CHROMEOS_DEVICE,
			ResourceState: dutstate.ConvertToUFSState(dut.State),
			DeviceRecoveryData: &ufsAPI.UpdateDeviceRecoveryDataRequest_Chromeos{
				Chromeos: &ufsAPI.ChromeOsRecoveryData{
					DutState: getUFSDutComponentStateFromSpecs(dutID, dut),
					DutData: &ufsAPI.ChromeOsRecoveryData_DutData{
						SerialNumber: devboard.GetSerialNumber(),
					},
					LabData: getUFSLabDataFromSpecs(dut),
				},
			},
		}, nil
	}
	return nil, errors.Reason("Unknown DUT type: %+v", dut).Err()
}

func adaptUfsDutToTLWDut(data *ufspb.ChromeOSDeviceData) (*tlw.Dut, error) {
	lc := data.GetLabConfig()
	dut := lc.GetChromeosMachineLse().GetDeviceLse().GetDut()
	p := dut.GetPeripherals()
	ds := data.GetDutState()
	dc := data.GetDeviceConfig()
	machine := data.GetMachine()
	make := data.GetManufacturingConfig()
	name := lc.GetName()
	var battery *tlw.Battery
	supplyType := tlw.ChromeOS_POWER_SUPPLY_UNSPECIFIED
	if dc != nil {
		switch dc.GetPower() {
		case ufsdevice.Config_POWER_SUPPLY_BATTERY:
			supplyType = tlw.ChromeOS_BATTERY
			battery = &tlw.Battery{
				State: convertHardwareState(ds.GetBatteryState()),
			}
		case ufsdevice.Config_POWER_SUPPLY_AC_ONLY:
			supplyType = tlw.ChromeOS_AC_ONLY
		}
	}
	setup := tlw.DUTSetupTypeCros
	// TODO(b/270274087): return DUT setup type from lab service directly
	if strings.Contains(name, "jetstream") {
		setup = tlw.DUTSetupTypeJetstream
	}
	if machine.GetChromeosMachine().GetModel() == "betty" {
		setup = tlw.DUTSetupTypeCrosVM
	}
	// Check hostname to see if it's DUTs for browser testing
	if strings.HasPrefix(name, "chrome-") || strings.HasPrefix(name, "chromium-") {
		setup = tlw.DUTSetupTypeCrosBrowser
	}

	audio := &tlw.DUTAudio{
		LoopbackState: convertAudioLoopbackState(ds.GetAudioLoopbackDongle()),
		InBox:         p.GetAudio().GetAudioBox(),
		StaticCable:   p.GetAudio().GetAudioCable(),
	}
	d := &tlw.Dut{
		Id:             machine.GetName(),
		Name:           name,
		SetupType:      setup,
		State:          dutstate.ConvertFromUFSState(lc.GetResourceState()),
		RepairRequests: convertRepairRequestsFromUFS(ds.GetRepairRequests()),
		Chromeos: &tlw.ChromeOS{
			Board:               machine.GetChromeosMachine().GetBuildTarget(),
			Model:               machine.GetChromeosMachine().GetModel(),
			Hwid:                machine.GetChromeosMachine().GetHwid(),
			SerialNumber:        machine.GetSerialNumber(),
			Phase:               make.GetDevicePhase().String()[len("PHASE_"):],
			PowerSupplyType:     supplyType,
			Audio:               audio,
			Servo:               createServoHost(p.GetServo(), p.GetSmartUsbhub(), ds),
			Cr50Phase:           convertCr50Phase(ds.GetCr50Phase()),
			Cr50KeyEnv:          convertCr50KeyEnv(ds.GetCr50KeyEnv()),
			DeviceSku:           machine.GetChromeosMachine().GetSku(),
			Storage:             createDUTStorage(dc, ds),
			Wifi:                createDUTWifi(make, ds),
			Bluetooth:           createDUTBluetooth(ds, dc),
			Cellular:            createDUTCellular(ds, p),
			Battery:             battery,
			Chameleon:           createChameleon(p, ds),
			WifiRouters:         createWifiRouterHosts(p.GetWifi()),
			PeripheralWifiState: convertPeripheralWifiState(ds.GetWifiPeripheralState()),
			BluetoothPeers:      createBluetoothPeerHosts(p),
			RpmOutlet:           createRPMOutlet(p.GetRpm(), ds),
			RoVpdMap:            dut.GetRoVpdMap(),
			Cbi:                 dut.GetCbi(),
			Cbx:                 dut.GetCbx(),
			HumanMotionRobot:    createDUTHumanMotionRobot(p, ds),
		},
		ExtraAttributes: map[string][]string{
			tlw.ExtraAttributePools: dut.GetPools(),
		},
		ProvisionedInfo: &tlw.ProvisionedInfo{},
	}
	d.Chromeos.WifiRouterFeatures = p.Wifi.GetWifiRouterFeatures()

	if p.GetServo().GetServoSetup() == ufslab.ServoSetupType_SERVO_SETUP_DUAL_V4 {
		d.ExtraAttributes[tlw.ExtraAttributeServoSetup] = []string{tlw.ExtraAttributeServoSetupDual}
	}
	if ds.GetDutStateReason() != "" {
		d.DutStateReason = tlw.DutStateReason(ds.GetDutStateReason())
	}
	return d, nil
}

// createBluetoothPeerHosts use the UFS states for Bluetooth peer devices to create
// the equivalent tlw slice.
func createBluetoothPeerHosts(peripherals *ufslab.Peripherals) []*tlw.BluetoothPeer {
	var bluetoothPeerHosts []*tlw.BluetoothPeer
	for _, btp := range peripherals.GetBluetoothPeers() {
		var (
			hostname string
			state    tlw.BluetoothPeer_State
		)
		switch d := btp.GetDevice().(type) {
		case *ufslab.BluetoothPeer_RaspberryPi:
			hostname = d.RaspberryPi.GetHostname()
			state = convertBluetoothPeerState(d.RaspberryPi.GetState())
		default:
			// We never want this to fail. It does create a risk
			// for silent errors however. Introduction of new device
			// types is very infrequent and also a very conscious
			// event, which helps counterweight that risk.
			continue
		}
		bluetoothPeerHosts = append(bluetoothPeerHosts, &tlw.BluetoothPeer{
			Name:  hostname,
			State: state,
		})
	}

	return bluetoothPeerHosts
}

func adaptUfsLabstationToTLWDut(data *ufspb.ChromeOSDeviceData) (*tlw.Dut, error) {
	lc := data.GetLabConfig()
	l := lc.GetChromeosMachineLse().GetDeviceLse().GetLabstation()
	ds := data.GetDutState()
	dc := data.GetDeviceConfig()
	machine := data.GetMachine()
	make := data.GetManufacturingConfig()
	name := lc.GetName()
	d := &tlw.Dut{
		Id:        machine.GetName(),
		Name:      name,
		SetupType: tlw.DUTSetupTypeLabstation,
		Chromeos: &tlw.ChromeOS{
			Board:           machine.GetChromeosMachine().GetBuildTarget(),
			Model:           machine.GetChromeosMachine().GetModel(),
			Hwid:            machine.GetChromeosMachine().GetHwid(),
			SerialNumber:    machine.GetSerialNumber(),
			Phase:           make.GetDevicePhase().String()[len("PHASE_"):],
			PowerSupplyType: tlw.ChromeOS_AC_ONLY,

			Cr50Phase:  convertCr50Phase(ds.GetCr50Phase()),
			Cr50KeyEnv: convertCr50KeyEnv(ds.GetCr50KeyEnv()),
			DeviceSku:  machine.GetChromeosMachine().GetSku(),
			Storage:    createDUTStorage(dc, ds),
			RpmOutlet:  createRPMOutlet(l.GetRpm(), ds),
		},
		ExtraAttributes: map[string][]string{
			tlw.ExtraAttributePools: l.GetPools(),
		},
		ProvisionedInfo: &tlw.ProvisionedInfo{},
	}
	if ds.GetDutStateReason() != "" {
		d.DutStateReason = tlw.DutStateReason(ds.GetDutStateReason())
	}
	return d, nil
}

func adaptUfsDevBoardToTLWDut(data *ufspb.ChromeOSDeviceData) (*tlw.Dut, error) {
	lc := data.GetLabConfig()
	db := lc.GetChromeosMachineLse().GetDeviceLse().GetDevboard()
	ds := data.GetDutState()
	machine := data.GetMachine()
	name := lc.GetName()
	var board string
	if andreiBoard := machine.GetDevboard().GetAndreiboard(); andreiBoard != nil {
		board = "andreiboard"
	} else if icetower := machine.GetDevboard().GetIcetower(); icetower != nil {
		board = "icetower"
	} else if dragonclaw := machine.GetDevboard().GetDragonclaw(); dragonclaw != nil {
		board = "dragonclaw"
	}
	d := &tlw.Dut{
		Id:        machine.GetName(),
		Name:      name,
		SetupType: tlw.DUTSetupTypeDevBoard,
		DevBoard: &tlw.DevBoard{
			Board:        board,
			Model:        board,
			SerialNumber: machine.GetSerialNumber(),
			Servo:        createServoHost(db.GetServo(), false, ds),
		},
		ExtraAttributes: map[string][]string{
			tlw.ExtraAttributePools: db.GetPools(),
		},
		ProvisionedInfo: &tlw.ProvisionedInfo{},
	}
	if db.GetServo().GetServoSetup() == ufslab.ServoSetupType_SERVO_SETUP_DUAL_V4 {
		d.ExtraAttributes[tlw.ExtraAttributeServoSetup] = []string{tlw.ExtraAttributeServoSetupDual}
	}
	return d, nil
}

func createRPMOutlet(rpm *ufslab.OSRPM, ds *ufslab.DutState) *tlw.RPMOutlet {
	if rpm == nil || rpm.GetPowerunitName() == "" || rpm.GetPowerunitOutlet() == "" {
		return &tlw.RPMOutlet{
			State: convertRPMState(ds.GetRpmState()),
		}
	}
	return &tlw.RPMOutlet{
		Hostname: rpm.GetPowerunitName(),
		Outlet:   rpm.GetPowerunitOutlet(),
		State:    convertRPMState(ds.GetRpmState()),
	}
}

func createServoHost(servo *ufslab.Servo, useSmartUsbhub bool, ds *ufslab.DutState) *tlw.ServoHost {
	if servo.GetServoHostname() == "" {
		return nil
	}
	return &tlw.ServoHost{
		Name:               servo.GetServoHostname(),
		UsbkeyState:        convertHardwareState(ds.GetServoUsbState()),
		ServodPort:         servo.GetServoPort(),
		State:              convertServoState(ds.GetServo()),
		SerialNumber:       servo.GetServoSerial(),
		FirmwareChannel:    convertFirmwareChannel(servo.GetServoFwChannel()),
		ServodType:         servo.GetServoType(),
		SmartUsbhubPresent: useSmartUsbhub,
		ServoTopology:      convertServoTopologyFromUFS(servo.GetServoTopology()),
		ContainerName:      servo.GetDockerContainerName(),
		UsbDrive:           servo.GetUsbDrive(),
	}
}

func createChameleon(p *ufslab.Peripherals, ds *ufslab.DutState) *tlw.Chameleon {
	pCham := p.GetChameleon()
	cham := &tlw.Chameleon{
		Name:                     pCham.GetHostname(),
		State:                    convertChameleonState(ds.GetChameleon()),
		Audioboxjackpluggerstate: convertAudioBoxJackPluggerState(pCham.GetAudioboxJackplugger()),
	}

	if rpm := pCham.GetRpm(); rpm != nil {
		cham.RPMOutlet = &tlw.RPMOutlet{
			Hostname: rpm.GetPowerunitName(),
			Outlet:   rpm.GetPowerunitOutlet(),
		}
	}
	return cham
}

// createDUTHumanMotionRobot convert ufslab.Peripherals.HumanMotionRobot to tlw.HumanMotionRobot
// It includes hostname and state, and will be used for recovery
func createDUTHumanMotionRobot(p *ufslab.Peripherals, ds *ufslab.DutState) *tlw.HumanMotionRobot {
	pHmr := p.GetHumanMotionRobot()
	tlwHmr := &tlw.HumanMotionRobot{
		Name:  pHmr.GetHostname(),
		State: convertHumanMotionRobotStateToTLW(ds.GetHmrState()),
	}
	return tlwHmr
}

func createDUTStorage(dc *ufsdevice.Config, ds *ufslab.DutState) *tlw.Storage {
	return &tlw.Storage{
		Type:  convertStorageType(dc.GetStorage()),
		State: convertHardwareState(ds.GetStorageState()),
	}
}

func createDUTWifi(make *ufsmake.ManufacturingConfig, ds *ufslab.DutState) *tlw.Wifi {
	return &tlw.Wifi{
		State:    convertHardwareState(ds.GetWifiState()),
		ChipName: make.GetWifiChip(),
	}
}

// createWifiRouterHosts convert ufslab.Wifi.WifiRouters to []*tlw.WifiRouterHost
// It include router hostname, model, board, state, rpm information which will be used to verification and recovery
func createWifiRouterHosts(wifi *ufslab.Wifi) []*tlw.WifiRouterHost {
	var routers []*tlw.WifiRouterHost
	for _, ufsRouter := range wifi.GetWifiRouters() {
		tlwRpm := tlw.RPMOutlet{
			// TODO(otabek) update when http://b/216315183 is done.
			//set to unknown till rpm is updated to enable peripherals.
			//currently,rpm only supports on dut. router rpm state is not defined in proto yet and no api for rpmoutlet for non dut
			State: convertRPMState(ufslab.PeripheralState_UNKNOWN),
		}
		if rpm := ufsRouter.GetRpm(); rpm != nil && rpm.GetPowerunitName() != "" && rpm.GetPowerunitOutlet() != "" {
			tlwRpm.Hostname = rpm.GetPowerunitName()
			tlwRpm.Outlet = rpm.GetPowerunitOutlet()
		}
		routers = append(routers, &tlw.WifiRouterHost{
			Name:       ufsRouter.GetHostname(),
			State:      convertWifiRouterState(ufsRouter.GetState()),
			Model:      ufsRouter.GetModel(),
			Board:      ufsRouter.GetBuildTarget(),
			RPMOutlet:  &tlwRpm,
			Features:   ufsRouter.GetFeatures(),
			DeviceType: ufsRouter.GetDeviceType(),
		})
	}
	return routers
}

func createDUTBluetooth(ds *ufslab.DutState, dc *ufsdevice.Config) *tlw.Bluetooth {
	return &tlw.Bluetooth{
		Expected: configHasFeature(dc, ufsdevice.Config_HARDWARE_FEATURE_BLUETOOTH),
		State:    convertHardwareState(ds.GetBluetoothState()),
	}
}

func createDUTCellular(ds *ufslab.DutState, p *ufslab.Peripherals) *tlw.Cellular {
	return &tlw.Cellular{
		ModemState: convertHardwareState(ds.GetCellularModemState()),
		Carrier:    p.GetCarrier(),
	}
}

func configHasFeature(dc *ufsdevice.Config, hf ufsdevice.Config_HardwareFeature) bool {
	for _, f := range dc.GetHardwareFeatures() {
		if f == hf {
			return true
		}
	}
	return false
}

func getUFSDutDataFromSpecs(dut *tlw.Dut) *ufsAPI.ChromeOsRecoveryData_DutData {
	dutData := &ufsAPI.ChromeOsRecoveryData_DutData{
		SerialNumber: dut.GetChromeos().GetSerialNumber(),
		HwID:         dut.GetChromeos().GetHwid(),
		// TODO: update logic if required by b/184391605
		DeviceSku: dut.GetChromeos().GetDeviceSku(),
	}
	return dutData
}

func getUFSLabDataFromSpecs(dut *tlw.Dut) *ufsAPI.ChromeOsRecoveryData_LabData {
	labData := &ufsAPI.ChromeOsRecoveryData_LabData{
		WifiRouters: []*ufsAPI.ChromeOsRecoveryData_WifiRouter{},
	}
	if ch := dut.GetChromeos(); ch != nil {
		if s := ch.GetServo(); s != nil {
			labData.ServoType = s.GetServodType()
			labData.SmartUsbhub = s.GetSmartUsbhubPresent()
			labData.ServoTopology = convertServoTopologyToUFS(s.GetServoTopology())
			labData.ServoUsbDrive = s.GetUsbDrive()
		}
		for _, router := range ch.GetWifiRouters() {
			labData.WifiRouters = append(labData.WifiRouters, &ufsAPI.ChromeOsRecoveryData_WifiRouter{
				Hostname:   router.GetName(),
				State:      convertWifiRouterStateToUFS(router.GetState()),
				Model:      router.GetModel(),
				Features:   router.GetFeatures(),
				DeviceType: router.GetDeviceType(),
			})
		}
		for _, btp := range ch.GetBluetoothPeers() {
			labData.BluetoothPeers = append(labData.BluetoothPeers, &ufsAPI.ChromeOsRecoveryData_BluetoothPeer{
				Hostname: btp.GetName(),
				State:    convertBluetoothPeerStateToUFS(btp.GetState()),
			})
		}
		labData.RoVpdMap = ch.GetRoVpdMap()
		labData.Cbi = ch.GetCbi()
		labData.AudioboxJackpluggerState = convertAudioBoxJackPluggerStateToUFS(ch.GetChameleon().GetAudioboxjackpluggerstate())
		labData.WifiRouterFeatures = ch.GetWifiRouterFeatures()
	} else if ch := dut.GetDevBoard(); ch != nil {
		if s := ch.GetServo(); s != nil {
			labData.ServoType = s.GetServodType()
			labData.SmartUsbhub = s.GetSmartUsbhubPresent()
			labData.ServoTopology = convertServoTopologyToUFS(s.GetServoTopology())
			labData.ServoUsbDrive = s.GetUsbDrive()
		}
	}
	return labData
}

// getUFSDutComponentStateFromSpecs collects all states for DUT and peripherals.
func getUFSDutComponentStateFromSpecs(dutID string, dut *tlw.Dut) *ufslab.DutState {
	state := &ufslab.DutState{
		Id:       &ufslab.ChromeOSDeviceID{Value: dutID},
		Hostname: dut.Name,
	}
	// Set all components states to default.
	// The state is updated later if component is present.
	state.Servo = ufslab.PeripheralState_MISSING_CONFIG
	state.ServoUsbState = ufslab.HardwareState_HARDWARE_UNKNOWN
	state.RpmState = ufslab.PeripheralState_MISSING_CONFIG
	state.StorageState = ufslab.HardwareState_HARDWARE_UNKNOWN
	state.BatteryState = ufslab.HardwareState_HARDWARE_UNKNOWN
	state.WifiState = ufslab.HardwareState_HARDWARE_UNKNOWN
	state.BluetoothState = ufslab.HardwareState_HARDWARE_UNKNOWN
	state.CellularModemState = ufslab.HardwareState_HARDWARE_UNKNOWN
	state.Chameleon = ufslab.PeripheralState_UNKNOWN
	state.HmrState = ufslab.PeripheralState_UNKNOWN
	state.WorkingBluetoothBtpeer = 0
	state.DutStateReason = string(dut.DutStateReason)
	state.RepairRequests = convertRepairRequestsToUFS(dut.RepairRequests)

	// Update states for present components.
	if chromeos := dut.GetChromeos(); chromeos != nil {
		if s := chromeos.GetServo(); s != nil {
			for us, ls := range servoStates {
				if ls == s.GetState() {
					state.Servo = us
				}
			}
			state.ServoUsbState = convertHardwareStateToUFS(s.GetUsbkeyState())
		}
		if rpm := chromeos.GetRpmOutlet(); rpm != nil {
			for us, ls := range rpmStates {
				if ls == rpm.GetState() {
					state.RpmState = us
				}
			}
		}
		for us, ls := range cr50Phases {
			if ls == chromeos.GetCr50Phase() {
				state.Cr50Phase = us
			}
		}
		for us, ls := range cr50KeyEnvs {
			if ls == chromeos.GetCr50KeyEnv() {
				state.Cr50KeyEnv = us
			}
		}
		if s := chromeos.GetStorage(); s != nil {
			state.StorageState = convertHardwareStateToUFS(s.GetState())
		}
		if b := chromeos.GetBattery(); b != nil {
			state.BatteryState = convertHardwareStateToUFS(b.GetState())
		}
		if w := chromeos.GetWifi(); w != nil {
			state.WifiState = convertHardwareStateToUFS(w.GetState())
		}
		if b := chromeos.GetBluetooth(); b != nil {
			state.BluetoothState = convertHardwareStateToUFS(b.GetState())
		}
		if c := chromeos.GetCellular(); c != nil {
			state.CellularModemState = convertHardwareStateToUFS(c.GetModemState())
		}
		if ch := chromeos.GetChameleon(); ch != nil {
			for us, rs := range chameleonStates {
				if ch.GetState() == rs {
					state.Chameleon = us
				}
			}
		}
		if hmr := chromeos.GetHumanMotionRobot(); hmr != nil {
			state.HmrState = convertHumanMotionRobotStateToUFS(hmr.GetState())
		}
		for _, btph := range chromeos.GetBluetoothPeers() {
			if btph.GetState() == tlw.BluetoothPeer_WORKING {
				state.WorkingBluetoothBtpeer += 1
			}
		}
		if chromeos.GetAudio().GetLoopbackState() == tlw.DUTAudio_LOOPBACK_WORKING {
			state.AudioLoopbackDongle = ufslab.PeripheralState_WORKING
		} else {
			state.AudioLoopbackDongle = ufslab.PeripheralState_UNKNOWN
		}
		state.WifiPeripheralState = convertPeripheralWifiStateToUFS(chromeos.GetPeripheralWifiState())
	} else if devboard := dut.GetDevBoard(); devboard != nil {
		if s := devboard.GetServo(); s != nil {
			for us, ls := range servoStates {
				if ls == s.GetState() {
					state.Servo = us
					break
				}
			}
			state.ServoUsbState = convertHardwareStateToUFS(s.GetUsbkeyState())
		}
	}
	return state
}
