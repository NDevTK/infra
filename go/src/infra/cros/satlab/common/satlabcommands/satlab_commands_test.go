// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package satlabcommands

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"infra/cros/satlab/common/utils/executor"
)

// TestGetHostIPShouldSuccess test `GetHostIP` function.
func TestGetHostIPShouldSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	expectedIP := "127.0.0.1"
	commandExecutor := &executor.FakeCommander{CmdOutput: expectedIP}

	res, err := GetHostIP(ctx, commandExecutor)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	if res != expectedIP {
		t.Errorf("Expected %v, got %v", expectedIP, res)
	}
}

// TestGetHostIPShouldFail test `GetHostIP` function.
func TestGetHostIPShouldFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	expectedError := errors.New("exec command failed")
	commandExecutor := &executor.FakeCommander{Err: expectedError}

	res, err := GetHostIP(ctx, commandExecutor)

	// Assert
	if err == nil {
		t.Errorf("Should return error, but got no error")
	}

	if res != "" {
		t.Errorf("Expected %v, got %v", "", res)
	}
}

// TestGetMacAddressShouldSuccess test `GetMacAddress` function.
func TestGetMacAddressShouldSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	hostname := "127.0.0.1"
	expectedMacAddress := "aa:bb:cc:dd:ee:ff"

	commandExecutor := &executor.FakeCommander{
		FakeFn: func(in *exec.Cmd) ([]byte, error) {
			cmd := strings.Join(in.Args, " ")
			if cmd == "/usr/local/bin/get_host_ip" {
				return []byte(hostname), nil
			} else if cmd == "/usr/local/bin/docker exec dhcp cat /sys/class/net/eth0/address" {
				return []byte(expectedMacAddress), nil
			}
			return nil, errors.New(fmt.Sprintf("handle command: %v", in.Path))
		},
		CmdOutput: fmt.Sprintf("%v/24 dev eth0 scope link  src %v", hostname, hostname),
	}

	res, err := GetMacAddress(ctx, commandExecutor)

	// Assert
	if err != nil {
		t.Errorf("Should not return error, but got an error: %v", err)
	}

	if res != expectedMacAddress {
		t.Errorf("Expected %v, got %v", expectedMacAddress, res)
	}
}

// TestGetMacAddressShouldFailWhenCommandExecutorFailed test `GetMacAddress` function.
func TestGetMacAddressShouldFailWhenCommandExecutorFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	expectedError := errors.New("exec command failed")
	commandExecutor := &executor.FakeCommander{Err: expectedError}

	res, err := GetMacAddress(ctx, commandExecutor)

	// Assert
	if err == nil {
		t.Errorf("Should return error, but got no error")
	}

	if res != "" {
		t.Errorf("Expected %v, got %v", "", res)
	}
}

// TestGetMacAddressShouldFailWhenGetNICNameFailed test `GetMacAddress` function.
func TestGetMacAddressShouldFailWhenGetNICNameFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	hostname := "127.0.0.1"
	macAddress := "aa:bb:cc:dd:ee:ff"

	commandExecutor := &executor.FakeCommander{
		FakeFn: func(in *exec.Cmd) ([]byte, error) {
			cmd := strings.Join(in.Args, " ")
			if cmd == "/usr/local/bin/get_host_ip" {
				return []byte(hostname), nil
			} else if cmd == "/usr/local/bin/docker exec dhcp cat /sys/class/net/eth0/address" {
				return []byte(macAddress), nil
			}
			return nil, errors.New(fmt.Sprintf("handle command: %v", in.Path))
		},
		CmdOutput: "",
	}

	res, err := GetMacAddress(ctx, commandExecutor)

	// Assert
	if err == nil {
		t.Errorf("Should return error, but got no error")
	}

	if res != "" {
		t.Errorf("Expected %v, got %v", "", res)
	}
}
