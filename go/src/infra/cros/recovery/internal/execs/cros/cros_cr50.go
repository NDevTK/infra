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
	"infra/cros/recovery/tlw"
)

const (
	// gsctool version command that used to check the RW and RO version
	cr50FWCmd = "gsctool -a -f"
	// findFWVersionRegexp is the regular expression for finding the RW/RO version from the output
	findFWVersionRegexp = `%s (\d+\.\d+\.\d+)`
	// findFWVersionRegexp is the regular expression for finding the RW/RO version from the output
	findFWKeyIdRegexp = `keyids:.*%s (\S+)`
)

// cr50FWComponent gets either the RW/RO firmware component from the output of the gsctool version cmd.
// fw component can be either version or keyid.
// @param region: "RW" or "RO"
// @param region: findFWVersionRegexp or findFWVersionRegexp
//
// @returns: Either the RO or RW of the FW component value
// Ex: 0.5.40 for fw version
//     0x87b73b67 for fw keyid
func cr50FWComponent(ctx context.Context, r execs.Runner, region tlw.CR50Region, findComponentRegexp string) (string, error) {
	output, err := r(ctx, cr50FWCmd)
	if err != nil {
		return "", errors.Annotate(err, "cr 50 fw component").Err()
	}
	log.Debug(ctx, "CR 50 FW Info: %s", output)
	componentRegexp, err := regexp.Compile(fmt.Sprintf(findComponentRegexp, region))
	if err != nil {
		return "", errors.Annotate(err, "cr 50 fw component").Err()
	}
	matches := componentRegexp.FindStringSubmatch(output)
	if len(matches) == 0 {
		return "", errors.Reason("cr 50 fw component: %s not found", region).Err()
	}
	if len(matches) != 2 {
		return "", errors.Reason("cr 50 rw version: cr 50 fw output is in wrong format").Err()
	}
	componentValue := matches[1]
	log.Debug(ctx, "Found %s FW component of value: %s", region, componentValue)
	return componentValue, nil
}
