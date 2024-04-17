// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"strings"

	testapi "go.chromium.org/chromiumos/config/go/test/api"
)

func IsAndroid(hw *testapi.LegacyHW) bool {
	return strings.Contains(
		strings.ToLower(hw.GetBoard()),
		strings.ToLower("pixel"),
	)
}

func IsCros(hw *testapi.LegacyHW) bool {
	return (!IsDevBoard(hw) && !IsAndroid(hw))
}

func IsDevBoard(hw *testapi.LegacyHW) bool {

	return strings.Contains(
		strings.ToLower(hw.GetBoard()),
		strings.ToLower("-devboard"),
	)
}
