// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package ctprequest will build and return a CTP request to be handled by the CTP
// BuildBucket builder.
package ctprequest

import (
	requestpb "go.chromium.org/chromiumos/infra/proto/go/test_platform"
)

type (
	CTPRequests []*requestpb.Request
)
