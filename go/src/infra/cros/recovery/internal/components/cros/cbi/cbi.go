// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// CBI corruption detection and repair logic. go/cbi-auto-recovery-dd
package cbi

import (
	"context"
	"regexp"
	"time"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/log"

	"go.chromium.org/luci/common/errors"
)

// CBILocation stores the port and address needed to reference CBI contents in
// EEPROM.
type CBILocation struct {
	port    string
	address string
}

const (
	locateCBICommand = "ectool locatechip"
	cbiChipType      = "0" // Maps to CBI in the `ectool locatechip` utility
	cbiIndex         = "0" // Gets the first CBI chip (there is only ever one) on the DUT.
	locateCBIRegex   = `Port:\s(\d+).*Address:\s(0x\w+)`
)

// GetCBILocation uses the `ectool locatechip` utility to get the CBILocation
// from the DUT. Will return an error if the DUT doesn't support CBI or if it
// wasn't able to reach the DUT.
func GetCBILocation(ctx context.Context, run components.Runner) (*CBILocation, error) {
	locateCBIOutput, err := run(ctx, time.Second*30, locateCBICommand, cbiChipType, cbiIndex)
	if err != nil {
		return nil, errors.Annotate(err, "Unable to determine if CBI is present on the DUT").Err()
	}

	cbiLocation, err := buildCBILocation(locateCBIOutput)
	if err != nil {
		return nil, err
	}

	log.Infof(ctx, "Found CBI contents on the DUT")
	return cbiLocation, err
}

// buildCBILocation creates a CBILocation struct from the text output of an
// `ectool locatechip` call. Will return an error if the locateCBIOutput doesn't
// contain both the address and the port needed to access the CBI contents.
func buildCBILocation(locateCBIOutput string) (*CBILocation, error) {
	r, _ := regexp.Compile(locateCBIRegex)
	match := r.FindStringSubmatch(locateCBIOutput)
	if len(match) != 3 {
		return nil, errors.Reason("No CBI contents were found on the DUT").Err()
	}
	return &CBILocation{
		port:    match[1],
		address: match[2],
	}, nil
}
