// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package server

import (
	"log"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var utils = serverUtils{}

// serverUtils groups all utility methods for the package.
type serverUtils struct{}

// firstLine extracts the first line from a multiline string.
func (*serverUtils) firstLine(s string) string {
	return strings.Split(s, "\n")[0]
}

// mapToCode maps common docker error messages to gRPC status codes.
func (*serverUtils) mapToCode(errMsg string) codes.Code {
	switch {
	// docker errors
	case strings.HasPrefix(errMsg, "Error: No such network"):
		return codes.NotFound
	case strings.HasPrefix(errMsg, "Error: No such container"):
		return codes.NotFound
	case strings.Contains(errMsg, "operation is not permitted"):
		return codes.PermissionDenied
	case strings.Contains(errMsg, "already exists"):
		return codes.AlreadyExists
	// podman errors
	case strings.HasPrefix(errMsg, "Error: error inspecting object: no such network"):
		return codes.NotFound
	case strings.HasPrefix(errMsg, "Error: error inspecting object: no such container"):
		return codes.NotFound
	case strings.Contains(errMsg, "is already used"):
		return codes.AlreadyExists
	default:
		log.Println("unable to map error message to a known code:", errMsg)
		return codes.Unknown
	}
}

// toStatusError converts stderr output string to gRPC status error
func (u *serverUtils) toStatusError(stderrOutput string) error {
	return u.toStatusErrorWithMapper(stderrOutput, u.mapToCode)
}

// toStatusErrorWithMapper converts stderr output string to gRPC status error using a custom code mapping function
func (u *serverUtils) toStatusErrorWithMapper(stderrOutput string, mapper func(string) codes.Code) error {
	log.Println("stderr:", stderrOutput)
	errMsg := strings.TrimSpace(stderrOutput)
	return status.Error(mapper(errMsg), errMsg)
}

// notFound returns an NotFound gRPC status error
func (*serverUtils) notFound(reason string) error {
	return status.Error(codes.NotFound, reason)
}

// invalidArgument returns an InvalidArgument gRPC status error
func (*serverUtils) invalidArgument(reason string) error {
	return status.Error(codes.InvalidArgument, reason)
}

// unimplemented returns an Unimplemented gRPC status error
func (*serverUtils) unimplemented(reason string) error {
	return status.Error(codes.Unimplemented, reason)
}
