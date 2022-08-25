// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package util contains utility functions
package util

import (
	bbpb "go.chromium.org/luci/buildbucket/proto"
)

// Check if a step is compile step.
// For the moment, we only check for the step name.
// In the future, we may want to check for step tag (crbug.com/1353978)
func IsCompileStep(step *bbpb.Step) bool {
	return step.GetName() == "compile"
}
