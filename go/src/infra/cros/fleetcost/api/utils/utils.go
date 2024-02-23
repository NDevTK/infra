// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"google.golang.org/genproto/googleapis/type/money"

	fleetcostpb "infra/cros/fleetcost/api"
)

// ToIndicatorType converts a string to an indicator.
func ToIndicatorType(x string) fleetcostpb.IndicatorType {
	out, _ := lookupValue(fleetcostpb.IndicatorType_value, x, "INDICATOR_TYPE")
	return fleetcostpb.IndicatorType(out)
}

// ToUSD converts a string on the command line to US dollars.
//
// Right now, it works by parsing a float. I'm not sure this is great, we
// really should be parsing the number as a big decimal without going through a
// float, but the resulting code for doing that, right now, ends up being more
// complex than is justified.
//
// I'm writing this comment partially to express my frustration at not knowing
// a simpler way to parse an arbitrary-precision big decimal from the command
// line (none of the stuff in math/big is an exact match) and partially to
// exhort my future self or other readers to replace this function with something
// better.
func ToUSD(x string) *money.Money {
	const billion = 1000 * 1000 * 1000
	var val float64
	if _, err := fmt.Sscanf(x, "%f", &val); err != nil {
		return nil
	}
	units := int64(val)
	// Extract the fractional part multiply by 1000 and round to the nearest integer.
	fPart := math.Round(1000 * (val - float64(units)))
	nanos := int32((billion / 1000) * fPart)
	return &money.Money{
		CurrencyCode: "USD",
		Units:        units,
		Nanos:        nanos,
	}
}

// ToCostCadence converts a string to a cost cadence.
func ToCostCadence(x string) fleetcostpb.CostCadence {
	out, _ := lookupValue(fleetcostpb.CostCadence_value, x, "COST_CADENCE")
	return fleetcostpb.CostCadence(out)
}

var errNotFound error = errors.New("item not found")

func lookupValue(m map[string]int32, key string, prefix string) (int32, error) {
	key = strings.ToUpper(key)
	prefix = strings.ToUpper(prefix)
	candidates := []string{
		key,
		fmt.Sprintf("%s_%s", prefix, key),
	}
	for _, candidate := range candidates {
		if res, ok := m[candidate]; ok {
			return res, nil
		}
	}
	return 0, errNotFound
}
