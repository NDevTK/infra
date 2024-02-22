// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"fmt"
	"strings"

	fleetcostpb "infra/cros/fleetcost/api"
)

// ToIndicatorType converts a string to an indicator.
func ToIndicatorType(x string) fleetcostpb.IndicatorType {
	x = strings.ToUpper(x)
	candidates := []string{
		x,
		fmt.Sprintf("INDICATOR_TYPE_%s", x),
	}
	for _, candidate := range candidates {
		if res, ok := fleetcostpb.IndicatorType_value[candidate]; ok {
			return fleetcostpb.IndicatorType(res)
		}
	}
	return fleetcostpb.IndicatorType_INDICATOR_TYPE_UNKNOWN
}
