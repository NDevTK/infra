// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error messages for error verification
var (
	InvalidTest  string = "Test name not present in the allowlist for public tests : %s"
	InvalidModel string = "cannnot run public tests on a private model : %s"
	InvalidBoard string = "cannnot run public tests on a private board : %s"
	InvalidImage string = "Image name not present in the allowlist for public tests : %s"
)

// IsNotFoundError checks if an error has code NOT_FOUND
func IsNotFoundError(err error) bool {
	s, ok := status.FromError(err)
	if ok && s.Code() == codes.NotFound {
		return true
	}
	return false
}

// IsInternalError checks if an error has code INTERNAL
func IsInternalError(err error) bool {
	s, ok := status.FromError(err)
	if ok && s.Code() == codes.Internal {
		return true
	}
	return false
}

// IsInvalidTest checks if an invalid testName is passed
func IsInvalidTest(err error) bool {
	s, ok := status.FromError(err)
	if ok && s.Code() == codes.InvalidArgument && strings.Contains(s.Message(), strings.Split(InvalidTest, ":")[0]) {
		return true
	}
	return false
}

// IsInvalidModel checks if an invalid model is passed
func IsInvalidModel(err error) bool {
	s, ok := status.FromError(err)
	if ok && s.Code() == codes.InvalidArgument && strings.Contains(s.Message(), strings.Split(InvalidModel, ":")[0]) {
		return true
	}
	return false
}

// IsInvalidBoard checks if an invalid model is passed
func IsInvalidBoard(err error) bool {
	s, ok := status.FromError(err)
	if ok && s.Code() == codes.InvalidArgument && strings.Contains(s.Message(), strings.Split(InvalidBoard, ":")[0]) {
		return true
	}
	return false
}

// IsInvalidImage checks if an invalid image is passed
func IsInvalidImage(err error) bool {
	s, ok := status.FromError(err)
	if ok && s.Code() == codes.InvalidArgument && strings.Contains(s.Message(), strings.Split(InvalidImage, ":")[0]) {
		return true
	}
	return false
}
