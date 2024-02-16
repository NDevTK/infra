// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package swarming

import (
	"fmt"
	"strconv"
	"strings"

	"infra/libs/skylab/inventory"
)

func init() {
	converters = append(converters, basicConverter)
	reverters = append(reverters, basicReverter)
}

func basicConverter(dims Dimensions, ls *inventory.SchedulableLabels) {
	if v := ls.GetBoard(); v != "" {
		dims["label-board"] = []string{v}
	}
	if v := ls.GetModel(); v != "" {
		dims["label-model"] = []string{v}
	}
	if v := ls.GetSku(); v != "" {
		dims["label-sku"] = []string{v}
	}
	if v := ls.GetHwidSku(); v != "" {
		dims["label-hwid_sku"] = []string{v}
	}
	if v := ls.GetDlmSkuId(); v != "" {
		dims["label-dlm_sku_id"] = []string{v}
	}
	if v := ls.GetBrand(); v != "" {
		dims["label-brand"] = []string{v}
	}
	if v := ls.GetPlatform(); v != "" {
		dims["label-platform"] = []string{v}
	}
	if v := ls.GetReferenceDesign(); v != "" {
		dims["label-reference_design"] = []string{v}
	}
	if v := ls.GetWifiChip(); v != "" {
		dims["label-wifi_chip"] = []string{v}
	}
	if v := ls.GetEcType(); v != inventory.SchedulableLabels_EC_TYPE_INVALID {
		dims["label-ec_type"] = []string{v.String()}
	}
	if v := ls.GetOsType(); v != inventory.SchedulableLabels_OS_TYPE_INVALID {
		dims["label-os_type"] = []string{v.String()}
	}
	if v := ls.GetPhase(); v != inventory.SchedulableLabels_PHASE_INVALID {
		dims["label-phase"] = []string{v.String()}
	}
	for _, v := range ls.GetVariant() {
		if v != "" {
			appendDim(dims, "label-variant", v)
		}
	}
	for _, c := range ls.GetHwidComponent() {
		if strings.HasPrefix(c, "cellular/") {
			if v := strings.Split(c, "/")[1]; v != "" {
				dims["label-cellular_modem"] = []string{v}
			}
		}
	}

	if ls.GetStability() {
		dims["label-device-stable"] = []string{"True"}
	}
}

func basicReverter(ls *inventory.SchedulableLabels, d Dimensions) Dimensions {
	d = assignLastStringValueAndDropKey(d, ls.Board, "label-board")
	d = assignLastStringValueAndDropKey(d, ls.Model, "label-model")
	d = assignLastStringValueAndDropKey(d, ls.Sku, "label-sku")
	d = assignLastStringValueAndDropKey(d, ls.HwidSku, "label-hwid_sku")
	d = assignLastStringValueAndDropKey(d, ls.DlmSkuId, "label-dlm_sku_id")
	d = assignLastStringValueAndDropKey(d, ls.Brand, "label-brand")
	d = assignLastStringValueAndDropKey(d, ls.Platform, "label-platform")
	d = assignLastStringValueAndDropKey(d, ls.ReferenceDesign, "label-reference_design")
	d = assignLastStringValueAndDropKey(d, ls.WifiChip, "label-wifi_chip")
	if v, ok := getLastStringValue(d, "label-cellular_modem"); ok {
		ls.HwidComponent = append(ls.HwidComponent, fmt.Sprintf("cellular/%s", v))
		delete(d, "label-cellular_modem")
	}

	if v, ok := getLastStringValue(d, "label-ec_type"); ok {
		if ec, ok := inventory.SchedulableLabels_ECType_value[v]; ok {
			*ls.EcType = inventory.SchedulableLabels_ECType(ec)
		}
		delete(d, "label-ec_type")
	}
	if v, ok := getLastStringValue(d, "label-os_type"); ok {
		if ot, ok := inventory.SchedulableLabels_OSType_value[v]; ok {
			*ls.OsType = inventory.SchedulableLabels_OSType(ot)
		}
		delete(d, "label-os_type")
	}
	if v, ok := getLastStringValue(d, "label-phase"); ok {
		if p, ok := inventory.SchedulableLabels_Phase_value[v]; ok {
			*ls.Phase = inventory.SchedulableLabels_Phase(p)
		}
		delete(d, "label-phase")
	}
	ls.Variant = append(ls.Variant, d["label-variant"]...)
	delete(d, "label-variant")

	d = assignLastBoolValueAndDropKey(d, ls.Stability, "label-device-stable")
	return d
}

// assignLastStringValueAndDropKey assign the last string value matching the key to `to`
// and drop the key
func assignLastStringValueAndDropKey(d Dimensions, to *string, key string) Dimensions {
	if v, ok := getLastStringValue(d, key); ok {
		*to = v
	}
	delete(d, key)
	return d
}

// getLastStringValue return the last string value matching the key
func getLastStringValue(d Dimensions, key string) (string, bool) {
	if vs, ok := d[key]; ok {
		if len(vs) > 0 {
			return vs[len(vs)-1], true
		}
		return "", false
	}
	return "", false
}

// assignLastBoolValueAndDropKey assign the last bool value matching the key to `to`
// and drop the key
func assignLastBoolValueAndDropKey(d Dimensions, to *bool, key string) Dimensions {
	if v, ok := getLastBoolValue(d, key); ok {
		*to = v
	}
	delete(d, key)
	return d
}

// getLastBoolValue return the last bool value matching the key
func getLastBoolValue(d Dimensions, key string) (bool, bool) {
	if s, ok := getLastStringValue(d, key); ok {
		return strings.ToLower(s) == "true", true
	}
	return false, false
}

// assignLastInt32ValueAndDropKey assign the last int32 value matching the key to `to`
// and drop the key
func assignLastInt32ValueAndDropKey(d Dimensions, to *int32, key string) Dimensions {
	if v, ok := getLastInt32Value(d, key); ok {
		*to = v
	}
	delete(d, key)
	return d
}

// getLastInt32Value return the last int32 value matching the key
func getLastInt32Value(d Dimensions, key string) (int32, bool) {
	if s, ok := getLastStringValue(d, key); ok {
		if c, err := strconv.ParseInt(s, 10, 32); err == nil {
			return int32(c), true
		}
		return int32(-1), false
	}
	return int32(-1), false
}

// assignLastIntValueAndDropKey assign the last int value matching the key to `to`
// and drop the key
func assignLastIntValueAndDropKey(d Dimensions, to *int, key string) Dimensions {
	if v, ok := getLastIntValue(d, key); ok {
		*to = v
	}
	delete(d, key)
	return d
}

// getLastIntValue return the last int value matching the key
func getLastIntValue(d Dimensions, key string) (int, bool) {
	if s, ok := getLastStringValue(d, key); ok {
		if c, err := strconv.Atoi(s); err == nil {
			return c, true
		}
	}
	return -1, false
}
