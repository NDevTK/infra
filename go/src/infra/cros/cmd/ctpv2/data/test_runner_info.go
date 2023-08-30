// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package data

import "go.chromium.org/chromiumos/config/go/test/api"

// TrRequest represents the what will become a TR request:
// Where it will be 1 request that contains 1 --> many tests.
// This request will be built up to contain all needed information to make a Tr(v2) request.
type TrRequest struct {
	Req *api.HWRequirements
	Tcs []*api.CTPTestCase
}
