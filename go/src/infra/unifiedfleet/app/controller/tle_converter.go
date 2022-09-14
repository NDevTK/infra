// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"go.chromium.org/chromiumos/config/go/payload"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"

	"infra/libs/fleet/boxster/swarming"
	ufspb "infra/unifiedfleet/api/v1/models"
	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
)

// fs is a temporary file system that holds the TleSources mapping file.
//
//go:embed tle_sources.jsonproto
var fs embed.FS

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
	labelAliases, err := swarming.GetLabelNames(dutAttr)
	if err != nil {
		return nil, err
	}
	labelName := dutAttr.GetId().GetValue()

	tleSource, err := getTleLabelMapping(labelName)
	if err != nil {
		logging.Warningf(ctx, "fail to find TLE label mapping: %s", err.Error())
		return nil, nil
	}

	switch tleSource.GetSourceType() {
	case ufspb.TleSourceType_TLE_SOURCE_TYPE_DUT_STATE:
		return constructTleLabels(tleSource, labelAliases, dutState)
	case ufspb.TleSourceType_TLE_SOURCE_TYPE_LAB_CONFIG:
		return constructTleLabels(tleSource, labelAliases, lse)
	default:
		return nil, fmt.Errorf("%s is not a valid label source", tleSource.GetSourceType())
	}
}

// constructTleLabels returns label values of a set of label names.
//
// constructTleLabels retrieves label values from a proto message based on a
// given path. For each given label name, a full label in the form of
// `${name}:val1,val2` is constructed and returned as part of an array.
func constructTleLabels(tleSource *ufspb.TleSource, labelAliases []string, pm proto.Message) ([]string, error) {
	valsArr, err := swarming.GetLabelValues(fmt.Sprintf("$.%s", tleSource.GetFieldPath()), pm)
	if err != nil {
		return nil, err
	}

	switch tleSource.GetConverterType() {
	case ufspb.TleConverterType_TLE_CONVERTER_TYPE_STANDARD:
		return swarming.FormLabels(labelAliases, strings.Join(valsArr, ","))
	default:
		return swarming.FormLabels(labelAliases, strings.Join(valsArr, ","))
	}
}

// getTleLabelMapping gets the predefined label mapping based on a label name.
func getTleLabelMapping(labelName string) (*ufspb.TleSource, error) {
	mapFile, err := fs.ReadFile("tle_sources.jsonproto")
	if err != nil {
		return nil, err
	}

	var tleMappings ufspb.TleSources
	err = jsonpb.Unmarshal(bytes.NewBuffer(mapFile), &tleMappings)
	if err != nil {
		return nil, err
	}

	for _, tleSource := range tleMappings.GetTleSources() {
		if tleSource.GetLabelName() == labelName {
			return tleSource, nil
		}
	}

	return nil, fmt.Errorf("no TLE label mapping found for %s", labelName)
}
