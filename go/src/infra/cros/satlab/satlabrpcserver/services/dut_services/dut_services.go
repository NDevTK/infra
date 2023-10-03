// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dut_services

import (
	"context"

	"infra/cros/satlab/satlabrpcserver/models"
)

// Device is a struct provides the information
// of the IP address and is connected.
type Device struct {
	IP          string
	MACAddress  string
	IsConnected bool
}

// IDUTServices provides the services that regulate the DUTs.
type IDUTServices interface {
	// RunCommandOnIP send the command to the DUT device and then get the result back
	RunCommandOnIP(ctx context.Context, IP, cmd string) (*models.SSHResult, error)

	// RunCommandOnIPs send the command to DUT devices and then get the result back
	RunCommandOnIPs(ctx context.Context, IPs []string, cmd string) []*models.SSHResult

	// GetConnectedIPs get the connected IPs from `dnsmasq.lease`
	// and then check the IPs are alive.
	GetConnectedIPs(ctx context.Context) ([]Device, error)
}
