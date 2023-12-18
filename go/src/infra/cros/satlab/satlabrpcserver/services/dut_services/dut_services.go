// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dut_services

import (
	"context"

	"infra/cros/satlab/common/enumeration"
	"infra/cros/satlab/satlabrpcserver/models"
)

// Device is a struct provides the information
// of the IP address and is connected.
type Device struct {
	IP           string
	MACAddress   string
	IsPingable   bool
	HasTestImage bool
	ServoSerial  string
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

	// GetBoard get the DUT's board
	GetBoard(ctx context.Context, address string) (string, error)

	// GetModel gets the DUT's model
	GetModel(ctx context.Context, address string) (string, error)

	// GetGSCSerialAndServoUSBCount gets the Cr50 or Ti50 port serial number on the DUT
	GetGSCSerialAndServoUSBCount(ctx context.Context, address string) (*GSCInfo, error)

	// GetServoSerial gets the Servo serial number for given DUT IP address.
	GetServoSerial(ctx context.Context, IP string, usbDevices []enumeration.USBDevice) (bool, string, error)

	// GetUSBDevicePaths returns all the USBDevices instance of plugged devices
	GetUSBDevicePaths(ctx context.Context) ([]enumeration.USBDevice, error)
}
