// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package interfaces

import (
	testapi "go.chromium.org/chromiumos/config/go/test/api"
)

type CTPv2Builder interface {
	BuildRequest() *testapi.CTPv2Request
}

type DynamicTRv2Builder interface {
	BuildRequest() (*testapi.CrosTestRunnerDynamicRequest, error)
}
