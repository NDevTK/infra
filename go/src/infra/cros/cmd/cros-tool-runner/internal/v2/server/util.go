// Copyright 2022 The Chromium OS Authors. All rights reserved.
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
type serverUtils struct {
}

// firstLine extracts the first line from a multiline string.
func (*serverUtils) firstLine(s string) string {
	return strings.Split(s, "\n")[0]
}

// contains check if an element string exists in a slice.
func (*serverUtils) contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// mapToCode maps an error message to a corresponding gRPC status code.
func (*serverUtils) mapToCode(errMsg string) codes.Code {
	switch {
	// TODO(mingkong): add "No such container"
	case strings.HasPrefix(errMsg, "Error: No such network"):
		return codes.NotFound
	case strings.Contains(errMsg, "operation is not permitted"):
		return codes.PermissionDenied
	// TODO(mingkong): make this match more specific
	case strings.Contains(errMsg, "already exists"):
		return codes.AlreadyExists
	default:
		log.Println("unable to map error message to a known code:", errMsg)
		return codes.Unknown
	}
}

// toStatusError converts stderr output string to gRPC status error
func (u *serverUtils) toStatusError(stderrOutput string) error {
	errMsg := u.firstLine(stderrOutput)
	log.Println("first line from stderr:", errMsg)
	return status.Error(u.mapToCode(errMsg), errMsg)
}
