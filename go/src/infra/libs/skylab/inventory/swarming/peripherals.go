// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package swarming

import (
	"fmt"
	"strconv"
	"strings"

	"infra/libs/skylab/inventory"

	"go.chromium.org/chromiumos/infra/proto/go/lab"
)

func init() {
	converters = append(converters, boolPeripheralsConverter)
	reverters = append(reverters, boolPeripheralsReverter)
	converters = append(converters, otherPeripheralsConverter)
	reverters = append(reverters, otherPeripheralsReverter)
}

func boolPeripheralsConverter(dims Dimensions, ls *inventory.SchedulableLabels) {
	p := ls.GetPeripherals()
	if p.GetAudioBoard() {
		dims["label-audio_board"] = []string{"True"}
	}
	if p.GetAudioBox() {
		dims["label-audio_box"] = []string{"True"}
	}
	if p.GetAudioCable() {
		dims["label-audio_cable"] = []string{"True"}
	}
	if p.GetAudioLoopbackDongle() {
		dims["label-audio_loopback_dongle"] = []string{"True"}
	}
	if p.GetCamerabox() {
		dims["label-camerabox"] = []string{"True"}
	}
	if p.GetChameleon() {
		dims["label-chameleon"] = []string{"True"}
	}
	if p.GetConductive() {
		dims["label-conductive"] = []string{"True"}
	}
	if p.GetHuddly() {
		dims["label-huddly"] = []string{"True"}
	}
	if p.GetMimo() {
		dims["label-mimo"] = []string{"True"}
	}
	if p.GetServo() {
		dims["label-servo"] = []string{"True"}
	}
	if p.GetStylus() {
		dims["label-stylus"] = []string{"True"}
	}
	if p.GetWificell() {
		dims["label-wificell"] = []string{"True"}
	}
	if p.GetRouter_802_11Ax() {
		dims["label-router_802_11ax"] = []string{"True"}
	}
}

func boolPeripheralsReverter(ls *inventory.SchedulableLabels, d Dimensions) Dimensions {
	p := ls.Peripherals
	d = assignLastBoolValueAndDropKey(d, p.AudioBoard, "label-audio_board")
	d = assignLastBoolValueAndDropKey(d, p.AudioBox, "label-audio_box")
	d = assignLastBoolValueAndDropKey(d, p.AudioCable, "label-audio_cable")
	d = assignLastBoolValueAndDropKey(d, p.AudioLoopbackDongle, "label-audio_loopback_dongle")
	d = assignLastBoolValueAndDropKey(d, p.Camerabox, "label-camerabox")
	d = assignLastBoolValueAndDropKey(d, p.Chameleon, "label-chameleon")
	d = assignLastBoolValueAndDropKey(d, p.Conductive, "label-conductive")
	d = assignLastBoolValueAndDropKey(d, p.Huddly, "label-huddly")
	d = assignLastBoolValueAndDropKey(d, p.Mimo, "label-mimo")
	d = assignLastBoolValueAndDropKey(d, p.Servo, "label-servo")
	d = assignLastBoolValueAndDropKey(d, p.Stylus, "label-stylus")
	d = assignLastBoolValueAndDropKey(d, p.Wificell, "label-wificell")
	d = assignLastBoolValueAndDropKey(d, p.Router_802_11Ax, "label-router_802_11ax")
	return d
}

func otherPeripheralsConverter(dims Dimensions, ls *inventory.SchedulableLabels) {
	p := ls.GetPeripherals()
	for _, v := range p.GetChameleonType() {
		appendDim(dims, "label-chameleon_type", v.String())
	}

	if invSState := p.GetServoState(); invSState != inventory.PeripheralState_UNKNOWN {
		if labSState, ok := lab.PeripheralState_name[int32(invSState)]; ok {
			dims["label-servo_state"] = []string{labSState}
		}
	}

	if chamState := p.GetChameleonState(); chamState != inventory.PeripheralState_UNKNOWN {
		if labSState, ok := lab.PeripheralState_name[int32(chamState)]; ok {
			dims["label-chameleon_state"] = []string{labSState}
		}
	}

	if invJackPluggerState := p.GetAudioboxJackpluggerState(); invJackPluggerState != inventory.Peripherals_AUDIOBOX_JACKPLUGGER_UNSPECIFIED {
		labJackPluggerState := invJackPluggerState.String() // AUDIOBOX_JACKPLUGGER_{ UNSPECIFIED, WORKING, ... }
		const plen = len("AUDIOBOX_JACKPLUGGER_")
		dims["label-audiobox_jackplugger_state"] = []string{labJackPluggerState[plen:]}
	}

	if invTRRSType := p.GetTrrsType(); invTRRSType != inventory.Peripherals_TRRS_TYPE_UNSPECIFIED {
		labTRRSType := invTRRSType.String() // TRRS_TYPE_{ CTIA, OMTP, ... }
		const plen = len("TRRS_TYPE_")
		dims["label-trrs_type"] = []string{labTRRSType[plen:]}
	}

	if hmrState := p.GetHmrState(); hmrState != inventory.PeripheralState_UNKNOWN {
		if labSState, ok := lab.PeripheralState_name[int32(hmrState)]; ok {
			dims["label-hmr_state"] = []string{labSState}
		}
	}

	n := p.GetWorkingBluetoothBtpeer()
	btpeers := make([]string, n)
	for i := range btpeers {
		btpeers[i] = fmt.Sprint(i + 1)
	}
	// Empty dimensions may cause swarming page to fail to load: crbug.com/1056285
	if len(btpeers) > 0 {
		dims["label-working_bluetooth_btpeer"] = btpeers
	}

	if facing := p.GetCameraboxFacing(); facing != inventory.Peripherals_CAMERABOX_FACING_UNKNOWN {
		dims["label-camerabox_facing"] = []string{facing.String()}
	}

	if light := p.GetCameraboxLight(); light != inventory.Peripherals_CAMERABOX_LIGHT_UNKNOWN {
		dims["label-camerabox_light"] = []string{light.String()}
	}

	for _, v := range p.GetServoComponent() {
		appendDim(dims, "label-servo_component", v)
	}

	hardwareStatePrefixLength := len("HARDWARE_")
	if servoUSBState := p.GetServoUsbState(); servoUSBState != inventory.HardwareState_HARDWARE_UNKNOWN {
		if usbState, ok := lab.HardwareState_name[int32(p.GetServoUsbState())]; ok {
			appendDim(dims, "label-servo_usb_state", usbState[hardwareStatePrefixLength:])
		}
	}

	if wifiState := p.GetWifiState(); wifiState != inventory.HardwareState_HARDWARE_UNKNOWN {
		if wState, ok := lab.HardwareState_name[int32(wifiState)]; ok {
			appendDim(dims, "label-wifi_state", wState[hardwareStatePrefixLength:])
		}
	}

	if bluetoothState := p.GetBluetoothState(); bluetoothState != inventory.HardwareState_HARDWARE_UNKNOWN {
		if btState, ok := lab.HardwareState_name[int32(bluetoothState)]; ok {
			appendDim(dims, "label-bluetooth_state", btState[hardwareStatePrefixLength:])
		}
	}
	if modemState := p.GetCellularModemState(); modemState != inventory.HardwareState_HARDWARE_UNKNOWN {
		if state, ok := lab.HardwareState_name[int32(modemState)]; ok {
			appendDim(dims, "label-cellular_modem_state", state[hardwareStatePrefixLength:])
		}
	}
	if peripheralWifiState := p.GetPeripheralWifiState(); peripheralWifiState != inventory.PeripheralState_UNKNOWN {
		if pwsState, ok := lab.PeripheralState_name[int32(peripheralWifiState)]; ok {
			dims["label-peripheral_wifi_state"] = []string{pwsState}
		}
	}
	for _, v := range p.GetWifiRouterFeatures() {
		appendDim(dims, "label-wifi_router_features", v.String())
	}
	for _, v := range p.GetWifiRouterModels() {
		appendDim(dims, "label-wifi_router_models", v)
	}
}

func otherPeripheralsReverter(ls *inventory.SchedulableLabels, d Dimensions) Dimensions {
	p := ls.Peripherals

	p.ChameleonType = make([]inventory.Peripherals_ChameleonType, len(d["label-chameleon_type"]))
	for i, v := range d["label-chameleon_type"] {
		if ct, ok := inventory.Peripherals_ChameleonType_value[v]; ok {
			p.ChameleonType[i] = inventory.Peripherals_ChameleonType(ct)
		}
	}
	delete(d, "label-chameleon_type")

	if chamStateName, ok := getLastStringValue(d, "label-chameleon_state"); ok {
		chamState := inventory.PeripheralState_UNKNOWN
		if ssIndex, ok := lab.PeripheralState_value[strings.ToUpper(chamStateName)]; ok {
			chamState = inventory.PeripheralState(ssIndex)
		}
		p.ChameleonState = &chamState
		delete(d, "label-chameleon_state")
	}

	if labJackPluggerName, ok := getLastStringValue(d, "label-audiobox_jackplugger_state"); ok {
		labJackPluggerState := "AUDIOBOX_JACKPLUGGER_" + labJackPluggerName
		if invJackPluggerVal, ok := inventory.Peripherals_AudioBoxJackPlugger_value[labJackPluggerState]; ok {
			invJackPluggerState := inventory.Peripherals_AudioBoxJackPlugger(invJackPluggerVal)
			p.AudioboxJackpluggerState = &invJackPluggerState
		}
		delete(d, "label-audiobox_jackplugger_state")
	}

	if labTRRSTypeName, ok := getLastStringValue(d, "label-trrs_type"); ok {
		labTRRSType := "TRRS_TYPE_" + labTRRSTypeName
		if invTRRSVal, ok := inventory.Peripherals_TRRSType_value[labTRRSType]; ok {
			invTRRSType := inventory.Peripherals_TRRSType(invTRRSVal)
			p.TrrsType = &invTRRSType
		}
		delete(d, "label-trrs_type")
	}

	if hmrStateName, ok := getLastStringValue(d, "label-hmr_state"); ok {
		hmrState := inventory.PeripheralState_UNKNOWN
		if ssIndex, ok := lab.PeripheralState_value[strings.ToUpper(hmrStateName)]; ok {
			hmrState = inventory.PeripheralState(ssIndex)
		}
		p.HmrState = &hmrState
		delete(d, "label-hmr_state")
	}

	if labSStateName, ok := getLastStringValue(d, "label-servo_state"); ok {
		servoState := inventory.PeripheralState_UNKNOWN
		if ssIndex, ok := lab.PeripheralState_value[strings.ToUpper(labSStateName)]; ok {
			servoState = inventory.PeripheralState(ssIndex)
		}
		p.ServoState = &servoState
		delete(d, "label-servo_state")
	}

	btpeers := d["label-working_bluetooth_btpeer"]
	max := 0
	for _, v := range btpeers {
		if i, err := strconv.Atoi(v); err == nil && i > max {
			max = i
		}
	}
	*p.WorkingBluetoothBtpeer = int32(max)
	delete(d, "label-working_bluetooth_btpeer")

	if facingName, ok := getLastStringValue(d, "label-camerabox_facing"); ok {
		if index, ok := inventory.Peripherals_CameraboxFacing_value[strings.ToUpper(facingName)]; ok {
			facing := inventory.Peripherals_CameraboxFacing(index)
			p.CameraboxFacing = &facing
		}
		delete(d, "label-camerabox_facing")
	}

	if lightName, ok := getLastStringValue(d, "label-camerabox_light"); ok {
		if index, ok := inventory.Peripherals_CameraboxLight_value[strings.ToUpper(lightName)]; ok {
			light := inventory.Peripherals_CameraboxLight(index)
			p.CameraboxLight = &light
		}
		delete(d, "label-camerabox_light")
	}

	p.ServoComponent = make([]string, len(d["label-servo_component"]))
	for i, v := range d["label-servo_component"] {
		p.ServoComponent[i] = v
	}
	delete(d, "label-servo_component")

	if servoUSBState, ok := getLastStringValue(d, "label-servo_usb_state"); ok {
		if labSStateVal, ok := lab.HardwareState_value["HARDWARE_"+strings.ToUpper(servoUSBState)]; ok {
			state := inventory.HardwareState(labSStateVal)
			p.ServoUsbState = &state
		}
		delete(d, "label-servo_usb_state")
	}

	if wifiState, ok := getLastStringValue(d, "label-wifi_state"); ok {
		if labSStateVal, ok := lab.HardwareState_value["HARDWARE_"+strings.ToUpper(wifiState)]; ok {
			state := inventory.HardwareState(labSStateVal)
			p.WifiState = &state
		}
		delete(d, "label-wifi_state")
	}
	if bluetoothState, ok := getLastStringValue(d, "label-bluetooth_state"); ok {
		if labSStateVal, ok := lab.HardwareState_value["HARDWARE_"+strings.ToUpper(bluetoothState)]; ok {
			state := inventory.HardwareState(labSStateVal)
			p.BluetoothState = &state
		}
		delete(d, "label-bluetooth_state")
	}
	if modemState, ok := getLastStringValue(d, "label-cellular_modem_state"); ok {
		if labSStateVal, ok := lab.HardwareState_value["HARDWARE_"+strings.ToUpper(modemState)]; ok {
			state := inventory.HardwareState(labSStateVal)
			p.CellularModemState = &state
		}
		delete(d, "label-cellular_modem_state")
	}

	if pwsStateName, ok := getLastStringValue(d, "label-peripheral_wifi_state"); ok {
		pwsState := inventory.PeripheralState_UNKNOWN
		if sIndex, ok := lab.PeripheralState_value[strings.ToUpper(pwsStateName)]; ok {
			pwsState = inventory.PeripheralState(sIndex)
		}
		p.PeripheralWifiState = &pwsState
		delete(d, "label-peripheral_wifi_state")
	}

	p.WifiRouterFeatures = make([]inventory.Peripherals_WifiRouterFeature, len(d["label-wifi_router_features"]))
	for i, v := range d["label-wifi_router_features"] {
		int32Value, ok := inventory.Peripherals_WifiRouterFeature_value[v]
		if !ok {
			// Could an int if the infra enum copy is out of sync, so try to parse it.
			intValue, err := strconv.Atoi(v)
			if err != nil {
				intValue = int(inventory.Peripherals_WIFI_ROUTER_FEATURE_INVALID.Number())
			}
			int32Value = int32(intValue)
		}
		p.WifiRouterFeatures[i] = inventory.Peripherals_WifiRouterFeature(int32Value)
	}
	delete(d, "label-wifi_router_features")

	p.WifiRouterModels = make([]string, len(d["label-wifi_router_models"]))
	for i, v := range d["label-wifi_router_models"] {
		p.WifiRouterModels[i] = v
	}
	delete(d, "label-wifi_router_models")

	return d
}
