// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"infra/libs/cros/lab_inventory/utils"
)

func checkDuplicatedString() (input chan string, result chan bool) {
	input = make(chan string)
	result = make(chan bool)
	set := map[string]bool{}
	go func() {
		defer close(result)

		for i := range input {
			_, existing := set[i]
			set[i] = true
			result <- existing
		}
	}()
	return
}

// Validate validates input requests and return error if it's not.
//
// All devices should have unique hostname and/or id.
// Doesn't allow mix of DUT and labstation in RPC request. They should be
// deployed separatedly.
func (r *AddCrosDevicesRequest) Validate() error {
	if r.Devices == nil || len(r.Devices) == 0 {
		return status.Errorf(codes.InvalidArgument, "no devices to add")
	}

	hostnameChecker, duplicatedHostname := checkDuplicatedString()
	defer close(hostnameChecker)
	idChecker, duplicatedID := checkDuplicatedString()
	defer close(idChecker)

	deviceTypes, dutType, labstationType := 0x0, 0x1, 0x2
	for _, d := range r.Devices {
		// Hostname is required.
		hostname := utils.GetHostname(d)
		if hostname == "" {
			return status.Errorf(codes.InvalidArgument, "Hostname is missing")
		}
		// Hostname must be unique.
		if hostnameChecker <- hostname; <-duplicatedHostname {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Duplicated hostname found: %s", hostname))
		}

		// ID is optional, but it must be unique if presents.
		id := d.GetId().GetValue()
		if id != "" {
			if idChecker <- id; <-duplicatedID {
				return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Duplicated id found: %s", id))
			}
		}
		switch {
		case d.GetDut() != nil:
			deviceTypes |= dutType
		case d.GetLabstation() != nil:
			deviceTypes |= labstationType
		}
		if deviceTypes == dutType|labstationType {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("DUT and labstation mixed in one request"))
		}
	}
	return nil
}

// Validate validates input requests and return error if it's not.
func (r *DeleteCrosDevicesRequest) Validate() error {
	if r.Ids == nil || len(r.Ids) == 0 {
		return status.Errorf(codes.InvalidArgument, "no devices to remove")
	}
	hostnameChecker, duplicatedHostname := checkDuplicatedString()
	defer close(hostnameChecker)
	idChecker, duplicatedID := checkDuplicatedString()
	defer close(idChecker)

	for _, id := range r.Ids {
		if _, ok := id.GetId().(*DeviceID_Hostname); ok {
			hostname := id.GetHostname()
			if hostnameChecker <- hostname; <-duplicatedHostname {
				return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Duplicated hostname found: %s", hostname))
			}
		} else {
			devID := id.GetChromeosDeviceId()
			if idChecker <- devID; <-duplicatedID {
				return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Duplicated id found: %s", devID))
			}
		}
	}
	return nil
}

// Validate validates getting requests must be non-empty and return error if
// it's not.
func (r *GetCrosDevicesRequest) Validate() error {
	if r.Ids == nil || len(r.Ids) == 0 {
		return status.Errorf(codes.InvalidArgument, "must specify device ID(s) to get")
	}
	return nil
}

// Validate validates input requests and return error if it's not.
func (r *UpdateCrosDevicesSetupRequest) Validate() error {
	if r.Devices == nil || len(r.Devices) == 0 {
		return status.Errorf(codes.InvalidArgument, "no devices to update")
	}
	// There must be no dupicated ID in the request.
	idChecker, duplicatedID := checkDuplicatedString()
	defer close(idChecker)

	for _, d := range r.Devices {
		id := d.GetId().GetValue()
		if idChecker <- id; <-duplicatedID {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Duplicated id found: %s", id))
		}
	}
	return nil
}

// Validate validates input requests and return error if it's not.
func (r *UpdateDutsStatusRequest) Validate() error {
	if r.States == nil || len(r.States) == 0 {
		return status.Errorf(codes.InvalidArgument, "no devices to update")
	}
	// There must be no dupicated ID in the request.
	idChecker, duplicatedID := checkDuplicatedString()
	defer close(idChecker)

	idWithStates := make(map[string]bool, len(r.States))
	for _, d := range r.States {
		id := d.GetId().GetValue()
		idWithStates[id] = true
		if idChecker <- id; <-duplicatedID {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Duplicated id found: %s", id))
		}
	}

	idChecker2, duplicatedID2 := checkDuplicatedString()
	for _, d := range r.GetDutMetas() {
		id := d.GetChromeosDeviceId()
		if idChecker2 <- id; <-duplicatedID2 {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Duplicated id found in meta : %s", id))
		}
	}

	for _, d := range r.GetDutMetas() {
		id := d.GetChromeosDeviceId()
		if _, ok := idWithStates[id]; !ok {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot update meta without valid dut states: %s", id))
		}
	}
	return nil
}
