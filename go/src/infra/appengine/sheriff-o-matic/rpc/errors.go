// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"google.golang.org/grpc/codes"

	"go.chromium.org/luci/grpc/appstatus"
)

// invalidArgumentError annotates err as having an invalid argument.
// The error message is shared with the requester as is.
//
// Note that this differs from FailedPrecondition. It indicates arguments
// that are problematic regardless of the state of the system
// (e.g., a malformed file name).
func invalidArgumentError(err error) error {
	return appstatus.Attachf(err, codes.InvalidArgument, "%s", err)
}
