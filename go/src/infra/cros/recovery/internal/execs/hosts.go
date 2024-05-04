// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package execs

import "go.chromium.org/luci/common/errors"

// GetDeviceName returns name of the devices requested by code.
func (ei *ExecInfo) GetDeviceName(code string) (string, error) {
	var name string
	switch code {
	case "servo_host":
		name = ei.GetChromeos().GetServo().GetName()
	case "chameleon":
		name = ei.GetChromeos().GetChameleon().GetName()
	case "dut":
		name = ei.GetDut().Name
	case "active":
		name = ei.GetActiveResource()
	default:
		return "", errors.Reason("get device name: device code not specified").Err()
	}
	if name == "" {
		return "", errors.Reason("get device name: name is empty").Err()
	}
	return name, nil
}
