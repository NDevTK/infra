// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dut_services

import (
	"context"

	"infra/cros/satlab/satlabrpcserver/utils"
)

// IDUTServices provides the services that regulate the DUTs.
type IDUTServices interface {
	// RunCommandOnIP send the command to the DUT device and then get the result back
	RunCommandOnIP(ctx context.Context, IP, cmd string) (*utils.SSHResult, error)

	// RunCommandOnIPs send the command to DUT devices and then get the result back
	RunCommandOnIPs(ctx context.Context, IPs []string, cmd string) []*utils.SSHResult
}
