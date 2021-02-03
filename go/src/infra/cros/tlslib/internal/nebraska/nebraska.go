// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package nebraska

import (
	"go.chromium.org/chromiumos/config/go/api/test/tls"
)

// Server represents a process of 'nebraska.py'.
type Server struct {
	Port int
}

// Start starts a nebraska server.
func Start(gsPathPrefix string, payloads []*tls.FakeOmaha_Payload, payloadsAddr string) (*Server, error) {
	// TODO (guocb): Add implementation.
	return &Server{}, nil
}

// Close terminates the nebraska server process and cleans up all temp
// dirs/files.
func (n *Server) Close() error {
	// TODO (guocb): Add implementation.
	return nil
}
