// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// CBI corruption detection and repair logic. go/cbi-auto-recovery-dd
package cbi

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"infra/cros/recovery/internal/components"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
)

// CBILocation stores the port and address needed to reference CBI contents in
// EEPROM.
type CBILocation struct {
	port    string
	address string
}

const (
	locateCBICommand   = "ectool locatechip"
	cbiChipType        = "0" // Maps to CBI in the `ectool locatechip` utility
	cbiIndex           = "0" // Gets the first CBI chip (there is only ever one) on the DUT.
	transferCBICommand = "ectool i2cxfer"
	cbiSize            = 256 // How many bytes of memory are stored in CBI.

	// How many bytes can be read from CBI in a single operation.
	// THIS VALUE SHOULD BE TREATED AS A HARD LIMIT. Exceeding this limit may
	// result in undefined behavior.
	readIncrement = 64
)

var readCBIRegex = regexp.MustCompile(`0x[[:xdigit:]]{1,2}|00`) // Match bytes printed in hex format (e.g. 00, 0x12, 0x3)
var locateCBIRegex = regexp.MustCompile(`Port:\s(\d+).*Address:\s(0x\w+)`)

// GetCBILocation uses the `ectool locatechip` utility to get the CBILocation
// from the DUT. Will return an error if the DUT doesn't support CBI or if it
// wasn't able to reach the DUT.
func GetCBILocation(ctx context.Context, run components.Runner) (*CBILocation, error) {
	locateCBIOutput, err := run(ctx, time.Second*30, locateCBICommand, cbiChipType, cbiIndex)
	if err != nil {
		return nil, errors.Annotate(err, "get CBI location: unable to determine if CBI is present on the DUT").Err()
	}

	cbiLocation, err := buildCBILocation(locateCBIOutput)
	return cbiLocation, errors.Annotate(err, "get CBI location").Err()
}

// buildCBILocation creates a CBILocation struct from the text output of an
// `ectool locatechip` call. Will return an error if the locateCBIOutput doesn't
// contain both the address and the port needed to access the CBI contents.
func buildCBILocation(locateCBIOutput string) (*CBILocation, error) {
	match := locateCBIRegex.FindStringSubmatch(locateCBIOutput)
	if len(match) < 3 {
		return nil, errors.Reason("build CBI location: no CBI contents were found on the DUT").Err()
	}
	return &CBILocation{
		port:    match[1],
		address: match[2],
	}, nil
}

// ReadCBIContents reads all <cbiSize> bytes from the CBI chip on the DUT in
// <readIncrement> sized reads using the ectool i2cxfer utility and returns a
// fully formed CBI proto.
func ReadCBIContents(ctx context.Context, run components.Runner, cbiLocation *CBILocation) (*labapi.Cbi, error) {
	hexContents := []string{}
	for offset := 0; offset < cbiSize; offset += readIncrement {
		cbiContents, err := run(ctx, time.Second*10, transferCBICommand, cbiLocation.port, cbiLocation.address, strconv.Itoa(readIncrement), strconv.Itoa(offset))
		if err != nil {
			return nil, errors.Annotate(err, "read CBI contents: unable to read CBI contents").Err()
		}

		hexBytes, err := parseBytesFromCBIContents(cbiContents, readIncrement)
		if err != nil {
			return nil, err
		}

		hexContents = append(hexContents, hexBytes...)
	}
	return &labapi.Cbi{RawContents: strings.Join(hexContents, " ")}, nil
}

// parseBytesFromCBIContents reads <numBytesToRead> number of bytes from the
// raw output from a call to `ectool i2cxfer` and returns a slice of bytes
// in hex format (the same format returned from `ectool i2cxfer`).
// e.g.
// cbiContents = "Read bytes: 0x43, 0x42, 0x49"
// numBytesToRead = 2
// hexBytes = ["0x43", "0x42"]
func parseBytesFromCBIContents(cbiContents string, numBytesToRead int) ([]string, error) {
	hexBytes := readCBIRegex.FindAllString(cbiContents, numBytesToRead)
	if len(hexBytes) != numBytesToRead {
		return nil, errors.Reason("parse bytes from CBI contents: wrong amount: expected %d bytes but read %d bytes instead. CBI contents found: %s", numBytesToRead, len(hexBytes), cbiContents).Err()
	}
	return hexBytes, nil
}
