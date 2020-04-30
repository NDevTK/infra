// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package deviceconfig

import (
	"fmt"
	"strings"

	"go.chromium.org/chromiumos/config/go/api"
	"go.chromium.org/chromiumos/config/go/payload"
	"go.chromium.org/chromiumos/infra/proto/go/device"
)

func parseConfigBundle(payload payload.ConfigBundle) []*device.Config {
	designs := payload.GetDesigns().GetValue()
	dcs := make(map[string]*device.Config, 0)
	for _, d := range designs {
		board := d.GetProgramId().GetValue()
		model := d.GetId().GetValue()
		for _, c := range d.GetConfigs() {
			dcs[c.GetId().GetValue()] = &device.Config{
				Id: &device.ConfigId{
					PlatformId: &device.PlatformId{Value: board},
					ModelId:    &device.ModelId{Value: model},
				},
				FormFactor:       parseFormFactor(c.GetHardwareFeatures().GetFormFactor().GetFormFactor()),
				HardwareFeatures: parseHardwareFeatures(payload.GetComponents(), c.GetHardwareFeatures()),
				Storage:          parseStorage(c.GetHardwareFeatures()),
				Cpu:              parseCPU(payload.GetComponents()),
				Soc:              parseSoc(payload.GetComponents()),
			}
		}
	}
	for _, sc := range payload.GetSoftwareConfigs() {
		designCID := sc.GetDesignConfigId().GetValue()
		v, ok := dcs[designCID]
		if !ok {
			continue
		}
		v.Id.VariantId = &device.VariantId{Value: fmt.Sprint(sc.GetIdScanConfig().GetFirmwareSku())}
	}
	res := make([]*device.Config, len(dcs))
	i := 0
	for _, v := range dcs {
		res[i] = v
		i++
	}
	return res
}

func parseFormFactor(ff api.HardwareFeatures_FormFactor_FormFactorType) device.Config_FormFactor {
	switch ff {
	case api.HardwareFeatures_FormFactor_CLAMSHELL:
		return device.Config_FORM_FACTOR_CLAMSHELL
	case api.HardwareFeatures_FormFactor_CONVERTIBLE:
		return device.Config_FORM_FACTOR_CONVERTIBLE
	case api.HardwareFeatures_FormFactor_DETACHABLE:
		return device.Config_FORM_FACTOR_DETACHABLE
	case api.HardwareFeatures_FormFactor_CHROMEBASE:
		return device.Config_FORM_FACTOR_CHROMEBASE
	case api.HardwareFeatures_FormFactor_CHROMEBOX:
		return device.Config_FORM_FACTOR_CHROMEBOX
	case api.HardwareFeatures_FormFactor_CHROMEBIT:
		return device.Config_FORM_FACTOR_CHROMEBIT
	case api.HardwareFeatures_FormFactor_CHROMESLATE:
		return device.Config_FORM_FACTOR_CHROMESLATE
	default:
		return device.Config_FORM_FACTOR_UNSPECIFIED
	}
}

func parseSoc(components *api.ComponentList) device.Config_SOC {
	for _, c := range components.GetValue() {
		if soc := c.GetSoc(); soc != nil {
			familyName := c.GetSoc().GetFamily().GetName()
			v, ok := device.Config_SOC_value[fmt.Sprintf("SOC_%s", strings.ToUpper(familyName))]
			if ok {
				return device.Config_SOC(v)
			}
		}
	}
	return device.Config_SOC_UNSPECIFIED
}

func parseHardwareFeatures(components *api.ComponentList, hf *api.HardwareFeatures) []device.Config_HardwareFeature {
	res := make([]device.Config_HardwareFeature, 0)
	for _, c := range components.GetValue() {
		switch c.GetType() {
		case &api.Component_Bluetooth_{}:
			res = append(res, device.Config_HARDWARE_FEATURE_BLUETOOTH)
		}
	}
	if hf.GetBluetooth() != nil {
		res = append(res, device.Config_HARDWARE_FEATURE_BLUETOOTH)
	}
	if hf.GetStylus() != nil {
		res = append(res, device.Config_HARDWARE_FEATURE_STYLUS)
	}
	if hf.GetFingerprint() != nil {
		res = append(res, device.Config_HARDWARE_FEATURE_FINGERPRINT)
	}
	if hf.GetKeyboard() != nil && hf.GetKeyboard().GetKeyboardType() == api.HardwareFeatures_Keyboard_DETACHABLE {
		res = append(res, device.Config_HARDWARE_FEATURE_DETACHABLE_KEYBOARD)
	}
	if screen := hf.GetScreen(); screen != nil {
		if screen.GetTouchSupport() == api.HardwareFeatures_PRESENT {
			res = append(res, device.Config_HARDWARE_FEATURE_TOUCHSCREEN)
		}
	}
	ff := hf.GetFormFactor().GetFormFactor()
	if ff != api.HardwareFeatures_FormFactor_CHROMEBOX && ff != api.HardwareFeatures_FormFactor_FORM_FACTOR_UNKNOWN {
		res = append(res, device.Config_HARDWARE_FEATURE_INTERNAL_DISPLAY)
	}
	if hf.GetCamera() != nil && hf.GetCamera().GetCount().GetValue() > 0 {
		res = append(res, device.Config_HARDWARE_FEATURE_WEBCAM)
	}
	return res
}

func parseStorage(hf *api.HardwareFeatures) device.Config_Storage {
	switch hf.GetStorage().GetStorageType() {
	case api.HardwareFeatures_Storage_NVME:
		return device.Config_STORAGE_NVME
	case api.HardwareFeatures_Storage_EMMC:
		return device.Config_STORAGE_MMC
	default:
		return device.Config_STORAGE_UNSPECIFIED
	}
}

func parseCPU(components *api.ComponentList) device.Config_Architecture {
	for _, c := range components.GetValue() {
		if soc := c.GetSoc(); soc != nil {
			switch soc.GetFamily().GetArch() {
			case api.Component_Soc_ARM:
				return device.Config_ARM
			case api.Component_Soc_ARM64:
				return device.Config_ARM64
			case api.Component_Soc_X86:
				return device.Config_X86
			case api.Component_Soc_X86_64:
				return device.Config_X86_64
			default:
				return device.Config_ARCHITECTURE_UNDEFINED
			}
		}
	}
	return device.Config_ARCHITECTURE_UNDEFINED
}
