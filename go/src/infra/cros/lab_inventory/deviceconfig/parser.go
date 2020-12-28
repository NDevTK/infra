// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package deviceconfig

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"go.chromium.org/chromiumos/config/go/api"
	"go.chromium.org/chromiumos/config/go/payload"
	"go.chromium.org/chromiumos/infra/proto/go/device"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	luciproto "go.chromium.org/luci/common/proto"

	"infra/libs/git"
)

var (
	unmarshaller = jsonpb.Unmarshaler{AllowUnknownFields: true}
)

type gitilesInfo struct {
	project string
	path    string
}

// Programs defines the structure of a DLM program list.
type Programs struct {
	Programs []struct {
		Name           string `json:"name,omitempty"`
		Repo           *Repo  `json:"repo,omitempty"`
		DeviceProjects []struct {
			Repo *Repo `json:"repo,omitempty"`
		} `json:"deviceProjects,omitempty"`
	} `json:"programs,omitempty"`
}

// Repo defines the repo info in DLM configs.
type Repo struct {
	Name       string `json:"name,omitempty"`
	RepoPath   string `json:"repoPath,omitempty"`
	ConfigPath string `json:"configPath,omitempty"`
}

func fixFieldMaskForConfigBundleList(b []byte) ([]byte, error) {
	var payload payload.ConfigBundleList
	t := reflect.TypeOf(payload)
	buf, err := luciproto.FixFieldMasksBeforeUnmarshal(b, t)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func getDeviceConfigs(ctx context.Context, gc git.ClientInterface, joinedConfigPath string) ([]*device.Config, error) {
	logging.Infof(ctx, "reading device configs from %s", joinedConfigPath)
	content, err := gc.GetFile(ctx, joinedConfigPath)
	if err != nil {
		return nil, err
	}
	var payloads payload.ConfigBundleList
	buf, err := fixFieldMaskForConfigBundleList([]byte(content))
	if err != nil {
		return nil, errors.Annotate(err, "fail to fix field mask for %s", joinedConfigPath).Err()
	}
	if err := unmarshaller.Unmarshal(bytes.NewBuffer(buf), &payloads); err != nil {
		return nil, errors.Annotate(err, "fail to unmarshal %s", joinedConfigPath).Err()
	}

	var allCfgs []*device.Config
	for _, payload := range payloads.GetValues() {
		dcs := parseConfigBundle(payload)
		allCfgs = append(allCfgs, dcs...)
	}
	return allCfgs, nil
}

func correctProjectName(n string) string {
	return strings.Replace(n, "+", "plus", -1)
}
func correctConfigPath(p string) string {
	return strings.Replace(p, "config.jsonproto", "joined.jsonproto", -1)
}

func validRepo(r *Repo) bool {
	return r != nil && r.Name != "" && r.ConfigPath != ""
}

func parseConfigBundle(configBundle *payload.ConfigBundle) []*device.Config {
	designs := configBundle.GetDesignList()
	dcs := make(map[string]*device.Config, 0)
	for _, d := range designs {
		board := d.GetProgramId().GetValue()
		model := d.GetName()
		for _, c := range d.GetConfigs() {
			dcs[c.GetId().GetValue()] = &device.Config{
				Id: &device.ConfigId{
					PlatformId: &device.PlatformId{Value: board},
					ModelId:    &device.ModelId{Value: model},
				},
				FormFactor: parseFormFactor(c.GetHardwareFeatures().GetFormFactor().GetFormFactor()),
				// TODO: GpuFamily, gpu_family in Component.Soc hasn't been set
				// Graphics: removed from boxster for now
				HardwareFeatures: parseHardwareFeatures(configBundle.GetComponents(), c.GetHardwareFeatures()),
				// TODO(xixuan): Power, a new power topology hasn't been set
				// label-power is used in swarming now: https://screenshot.googleplex.com/8EAUwGeoVeBtez7
				// TODO(xixuan): storage may be not accurate
				// label-storage is not used for scheduling tests for at least 3 months: https://screenshot.googleplex.com/B8spRMj22aUWkbb
				Storage: parseStorage(c.GetHardwareFeatures()),
				// TODO(xixuan): VideoAccelerationSupports, a new video acceleration topology hasn't been set
				// label-video_acceleration is not used for scheduling tests for at least 3 months: https://screenshot.googleplex.com/86h2scqNsStwoiW
				Soc: parseSoc(configBundle.GetComponents()),
				Cpu: parseArchitecture(configBundle.GetComponents()),
				Ec:  parseEcType(c.GetHardwareFeatures()),
			}
		}
	}
	// Setup the sku
	for _, sc := range configBundle.GetSoftwareConfigs() {
		designCID := sc.GetDesignConfigId().GetValue()
		dcs[designCID].Id.VariantId = &device.VariantId{Value: fmt.Sprint(sc.GetIdScanConfig().GetFirmwareSku())}
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

func parseSoc(components []*api.Component) device.Config_SOC {
	for _, c := range components {
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

func parseHardwareFeatures(components []*api.Component, hf *api.HardwareFeatures) []device.Config_HardwareFeature {
	res := make([]device.Config_HardwareFeature, 0)
	// Use bluetooth component & hardware feature to check
	if hf.GetBluetooth() != nil && hf.GetBluetooth().GetComponent() != nil {
		res = append(res, device.Config_HARDWARE_FEATURE_BLUETOOTH)
	} else {
		for _, c := range components {
			if c.GetBluetooth() != nil {
				res = append(res, device.Config_HARDWARE_FEATURE_BLUETOOTH)
				// May have multiple bluetooth, skip the following checks if bluetooth is already set.
				break
			}
		}
	}

	// TODO: HARDWARE_FEATURE_FLASHROM, not used
	// TODO: HARDWARE_FEATURE_HOTWORDING, field in topology.Audio hasn't been set
	// HARDWARE_FEATURE_INTERNAL_DISPLAY: Only chromeboxes have this UNSET
	ff := hf.GetFormFactor().GetFormFactor()
	if ff != api.HardwareFeatures_FormFactor_CHROMEBOX && ff != api.HardwareFeatures_FormFactor_FORM_FACTOR_UNKNOWN {
		res = append(res, device.Config_HARDWARE_FEATURE_INTERNAL_DISPLAY)
	}
	// TODO: HARDWARE_FEATURE_LUCID_SLEEP, which key in powerConfig?
	// HARDWARE_FEATURE_WEBCAM: hw_topo.create_camera
	// Ensure camera is not an empty object, e.g. "camera": {}
	if hf.GetCamera() != nil && hf.GetCamera().GetDevices() != nil {
		res = append(res, device.Config_HARDWARE_FEATURE_WEBCAM)
	}
	// Ensure stylus is not an empty object, e.g. "stylus": {}
	if hf.GetStylus() != nil {
		switch hf.GetStylus().GetStylus() {
		case api.HardwareFeatures_Stylus_STYLUS_UNKNOWN, api.HardwareFeatures_Stylus_NONE:
		default:
			res = append(res, device.Config_HARDWARE_FEATURE_STYLUS)
		}
	}
	// HARDWARE_FEATURE_TOUCHPAD: a component
	for _, c := range components {
		if c.GetTouchpad() != nil {
			res = append(res, device.Config_HARDWARE_FEATURE_TOUCHPAD)
			// May have multiple touchpads, skip the following checks if touchpad is already set.
			break
		}
	}
	// HARDWARE_FEATURE_TOUCHSCREEN: hw_topo.create_screen(touch=True)
	if screen := hf.GetScreen(); screen != nil {
		if screen.GetTouchSupport() == api.HardwareFeatures_PRESENT {
			res = append(res, device.Config_HARDWARE_FEATURE_TOUCHSCREEN)
		}
	}
	if fp := hf.GetFingerprint(); fp != nil {
		if fp.GetLocation() != api.HardwareFeatures_Fingerprint_NOT_PRESENT {
			res = append(res, device.Config_HARDWARE_FEATURE_FINGERPRINT)
		}
	}
	if hf.GetKeyboard() != nil && hf.GetKeyboard().GetKeyboardType() == api.HardwareFeatures_Keyboard_DETACHABLE {
		res = append(res, device.Config_HARDWARE_FEATURE_DETACHABLE_KEYBOARD)
	}
	return res
}

func parseStorage(hf *api.HardwareFeatures) device.Config_Storage {
	// TODO: How about other storage type?
	// STORAGE_SSD
	// STORAGE_HDD
	// STORAGE_UFS
	switch hf.GetStorage().GetStorageType() {
	case api.Component_Storage_NVME:
		return device.Config_STORAGE_NVME
	case api.Component_Storage_EMMC:
		return device.Config_STORAGE_MMC
	default:
		return device.Config_STORAGE_UNSPECIFIED
	}
}

func parseArchitecture(components []*api.Component) device.Config_Architecture {
	for _, c := range components {
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

func parseEcType(hf *api.HardwareFeatures) device.Config_EC {
	switch hf.GetEmbeddedController().GetEcType() {
	case api.HardwareFeatures_EmbeddedController_EC_CHROME:
		return device.Config_EC_CHROME
	case api.HardwareFeatures_EmbeddedController_EC_WILCO:
		return device.Config_EC_WILCO
	}
	return device.Config_EC_UNSPECIFIED
}
