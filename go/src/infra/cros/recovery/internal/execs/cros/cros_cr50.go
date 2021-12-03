// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"regexp"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

// CR50Region describes either the RO or RW of the firmware.
type CR50Region string

const (
	CR50RegionRO CR50Region = "RO"
	CR50RegionRW CR50Region = "RW"
)

const (
	// findFWVersionRegexp is the regular expression for finding the RW/RO version from the output
	findFWVersionRegexp = `%s (\d+\.\d+\.\d+)`
)

// GetCr50FwVersion gets the cr 50 firmware RO/RW version based on the region parameter.
// @param region: "RW" or "RO"
func GetCr50FwVersion(ctx context.Context, r execs.Runner, region CR50Region) (string, error) {
	fwVersion, err := cr50FWComponent(ctx, r, region, findFWVersionRegexp)
	if err != nil {
		return "", errors.Annotate(err, "get cr50 fw version").Err()
	}
	return fwVersion, nil
}

const (
	// gsctool version command that used to check the RW and RO version
	cr50FWCmd = "gsctool -a -f"
)

// cr50FWComponent gets either the RW/RO firmware component from the output of the gsctool version cmd.
// fw component can be either version or keyid.
// @param findComponentRegexp: findFWVersionRegexp or findFWVersionRegexp
//
// @returns: Either the RO or RW of the FW component value
// Ex: 0.5.40 for fw version
//     0x87b73b67 for fw keyid
func cr50FWComponent(ctx context.Context, r execs.Runner, region CR50Region, findComponentRegexp string) (string, error) {
	output, err := r(ctx, cr50FWCmd)
	if err != nil {
		return "", errors.Annotate(err, "cr50").Err()
	}
	log.Debug(ctx, "Cr50 : %s", output)
	componentRegexp, err := regexp.Compile(fmt.Sprintf(findComponentRegexp, region))
	if err != nil {
		return "", errors.Annotate(err, "cr50").Err()
	}
	matches := componentRegexp.FindStringSubmatch(output)
	if len(matches) == 0 {
		return "", errors.Reason("cr50: %s not found", region).Err()
	}
	if len(matches) != 2 {
		return "", errors.Reason("cr50: cr50 output is in wrong format").Err()
	}
	componentValue := matches[1]
	log.Debug(ctx, "Found %s FW component of value: %s", region, componentValue)
	return componentValue, nil
}
