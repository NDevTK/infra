// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/chromiumos/config/go/payload"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"

	"infra/libs/fleet/boxster/swarming"
	ufspb "infra/unifiedfleet/api/v1/models"
	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
)

const (
	dutStateSourceStr  = "dut_state"
	labConfigSourceStr = "lab_config"
)

type TleSource struct {
	configType string
	path       string
}

var TLE_LABEL_MAPPING = map[string]*TleSource{
	"attr-cr50-phase":             createTleSourceForConfig(dutStateSourceStr, "cr50_phase"),
	"attr-cr50-key-env":           createTleSourceForConfig(dutStateSourceStr, "cr50_key_env"),
	"attr-dut-id":                 createTleSourceForConfig(labConfigSourceStr, "name"),
	"attr-dut-name":               createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.hostname"),
	"hwid":                        createTleSourceForConfig(labConfigSourceStr, "UNIMPLEMENTED"),
	"serial_number":               createTleSourceForConfig(labConfigSourceStr, "UNIMPLEMENTED"),
	"misc-license":                createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.licenses..type"),
	"peripheral-arc":              createTleSourceForConfig(labConfigSourceStr, "UNIMPLEMENTED"),
	"peripheral-atrus":            createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.peripherals.audio.atrus"),
	"peripheral-audio-board":      createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.peripherals.chameleon.audio_board"),
	"peripheral-audio-box":        createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.peripherals.audio.audio_box"),
	"peripheral-audio-cable":      createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.peripherals.audio.audio_cable"),
	"peripheral-audio-loopback":   createTleSourceForConfig(labConfigSourceStr, "UNIMPLEMENTED"),
	"peripheral-bluetooth-state":  createTleSourceForConfig(dutStateSourceStr, "bluetooth_state"),
	"peripheral-camerabox-facing": createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.peripherals.camerabox_info.facing"),
	"peripheral-camerabox-light":  createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.peripherals.camerabox_info.light"),
	"peripheral-carrier":          createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.peripherals.carrier"),
	"peripheral-chameleon":        createTleSourceForConfig(labConfigSourceStr, "UNIMPLEMENTED"),
	"peripheral-chaos":            createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.peripherals.chaos"),
	"peripheral-mimo":             createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.peripherals.touch.mimo"),
	"peripheral-num-btpeer":       createTleSourceForConfig(dutStateSourceStr, "working_bluetooth_btpeer"),
	"peripheral-servo":            createTleSourceForConfig(labConfigSourceStr, "UNIMPLEMENTED"),
	"peripheral-servo-component":  createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.peripherals.servo..servo_component[*]"),
	"peripheral-servo-state":      createTleSourceForConfig(dutStateSourceStr, "servo"),
	"peripheral-servo-usb-state":  createTleSourceForConfig(dutStateSourceStr, "servo_usb_state"),
	"peripheral-wifi-state":       createTleSourceForConfig(dutStateSourceStr, "wifi_state"),
	"peripheral-wificell":         createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.peripherals.wifi.wificell"),
	"swarming-pool":               createTleSourceForConfig(labConfigSourceStr, "chromeos_machine_lse.device_lse.dut.pools"),
}

func createTleSourceForConfig(configType, path string) *TleSource {
	return &TleSource{
		configType: configType,
		path:       path,
	}
}

// Convert converts one DutAttribute label to multiple Swarming labels.
//
// For all TleSource labels needed to be converted for UFS, the implementation
// is handled in this file. All other labels uses the Boxster Swarming lib for
// conversion.
func Convert(ctx context.Context, dutAttr *api.DutAttribute, flatConfig *payload.FlatConfig, lse *ufspb.MachineLSE, dutState *chromeosLab.DutState) ([]string, error) {
	if dutAttr.GetTleSource() != nil {
		return convertTleSource(ctx, dutAttr, lse, dutState)
	}
	return swarming.ConvertAll(dutAttr, flatConfig)
}

// convertTleSource handles the label conversion of MachineLSE and DutState.
func convertTleSource(ctx context.Context, dutAttr *api.DutAttribute, lse *ufspb.MachineLSE, dutState *chromeosLab.DutState) ([]string, error) {
	labelNames, err := swarming.GetLabelNames(dutAttr)
	if err != nil {
		return nil, err
	}

	labelMapping, err := getTleLabelMapping(dutAttr.GetId().GetValue())
	if err != nil {
		logging.Warningf(ctx, "fail to find TLE label mapping: %s", err.Error())
		return nil, nil
	}

	switch labelMapping.configType {
	case dutStateSourceStr:
		return constructTleLabels(labelNames, labelMapping.path, dutState)
	case labConfigSourceStr:
		return constructTleLabels(labelNames, labelMapping.path, lse)
	default:
		return nil, fmt.Errorf("%s is not a valid label source", labelMapping.configType)
	}
}

// constructTleLabels returns label values of a set of label names.
//
// constructTleLabels retrieves label values from a proto message based on a
// given path. For each given label name, a full label in the form of
// `${name}:val1,val2` is constructed and returned as part of an array.
func constructTleLabels(labelNames []string, path string, pm proto.Message) ([]string, error) {
	valuesStr, err := swarming.GetLabelValuesStr(fmt.Sprintf("$.%s", path), pm)
	if err != nil {
		return nil, err
	}
	return swarming.FormLabels(labelNames, valuesStr)
}

// getTleLabelMapping gets the predefined label mapping based on a label name.
func getTleLabelMapping(label string) (*TleSource, error) {
	if val, ok := TLE_LABEL_MAPPING[label]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("no TLE label mapping found for %s", label)
}
