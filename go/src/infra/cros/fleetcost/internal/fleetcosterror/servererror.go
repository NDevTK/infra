// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package fleetcosterror provides error-related utilities,
// primarily for handling the issue of 500 errors being produced
// by *bare* Go errors.
package fleetcosterror

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WithDefaultCode supplies a default code.
func WithDefaultCode(code codes.Code, err error) error {
	if err == nil {
		return err
	}
	oldCode := status.Code(err)
	if oldCode == codes.Unknown {
		return status.Errorf(code, "%s", err)
	}
	return err
}
