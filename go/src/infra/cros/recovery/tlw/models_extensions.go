// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tlw

import (
	"strings"
)

// GetServodVersion gets a servod version, which is like a servod type but less specific.
func (x *ServoHost) GetServodVersion() string {
	servodType := x.GetServodType()
	// The v4p1 check must come first because v4 is a proper prefix of this prefix.
	if strings.HasPrefix(servodType, "servo_v4p1") {
		return "v4p1"
	}
	if strings.HasPrefix(servodType, "servo_v4") {
		return "v4"
	}
	return ""
}
