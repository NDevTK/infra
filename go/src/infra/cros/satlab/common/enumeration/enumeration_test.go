// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package enumeration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func setupTempFiles() (string, error) {
	//create temporary directorycfor testing
	dir, err := os.MkdirTemp("", "test-usb.device.path")
	if err != nil {
		return "", err
	}

	//create files within that directory
	vendorFile := filepath.Join(dir, "idVendor")
	if err := os.WriteFile(vendorFile, []byte("18d1"), 0666); err != nil {
		return "", err
	}
	productFile := filepath.Join(dir, "idProduct")
	if err := os.WriteFile(productFile, []byte("520d"), 0666); err != nil {
		return "", err
	}
	serialFile := filepath.Join(dir, "serial")
	if err := os.WriteFile(serialFile, []byte("servo-serial-1234"), 0666); err != nil {
		return "", err
	}

	return dir, nil
}

func TestNewUSBDevice(t *testing.T) {
	dir, err := setupTempFiles()
	if err != nil {
		t.Error("issues encountered when setting up test files")
	}
	defer func() {
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	dirPath := dir + "/"
	splitPath := strings.Split(dir, ".")
	hubPath := strings.Join(splitPath[:len(splitPath)-1], ".")

	cases := []struct {
		input  string
		output *USBDevice
		err    error
	}{
		{
			dirPath,
			&USBDevice{
				Serial:     "servo-serial-1234",
				DevicePath: dirPath,
				HubPath:    hubPath,
				DeviceType: "servo4.1",
			},
			nil,
		},
	}

	for _, tt := range cases {
		expected := tt.output
		actual, err := NewUSBDevice(tt.input)

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("expected: %v, unexpected diff: %s", expected, diff)
		}

		if !errors.Is(err, tt.err) {
			t.Errorf("expected: %v, got: %v", tt.err, err)
		}
	}
}

func TestGetPathContent(t *testing.T) {
	//create file for test
	f, err := os.CreateTemp("", "idVendor")
	if err != nil {
		t.Errorf("failed to create file for testing")
	}
	defer func() {
		err := os.Remove(f.Name())
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	if _, err := f.WriteString("18d1"); err != nil {
		t.Errorf("failed to create file for testing")
	}

	cases := []struct {
		input  string
		output string
		err    error
	}{
		{
			f.Name(),
			"18d1",
			nil,
		},
	}

	for _, tt := range cases {
		expected := tt.output
		actual, err := getPathContent(tt.input)

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("expected: %v, unexpected diff: %s", expected, diff)
		}

		if !errors.Is(err, tt.err) {
			t.Errorf("expected: %v, got: %v", tt.err, err)
		}
	}
}

func TestGetDevicesFromPaths(t *testing.T) {
	dir, err := setupTempFiles()
	if err != nil {
		t.Error("issues encountered when setting up test files")
	}
	defer func() {
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	dirPath := dir + "/"
	splitPath := strings.Split(dir, ".")
	hubPath := strings.Join(splitPath[:len(splitPath)-1], ".")

	cases := []struct {
		input  []string
		output []USBDevice
		err    error
	}{
		{
			[]string{dirPath},
			[]USBDevice{
				{
					Serial:     "servo-serial-1234",
					DevicePath: dirPath,
					HubPath:    hubPath,
					DeviceType: "servo4.1",
				},
			},
			nil,
		},
	}

	for _, tt := range cases {
		expected := tt.output
		actual, err := getDevicesFromPaths(tt.input)

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("expected: %v, unexpected diff: %s", expected, diff)
		}

		if !errors.Is(err, tt.err) {
			t.Errorf("expected: %v, got: %v", tt.err, err)
		}
	}
}

func TestFindServoFromDUT(t *testing.T) {
	cases := []struct {
		inputSerial     string
		inputDeviceList []USBDevice
		output          USBDevice
		err             error
	}{
		{
			"dut-serial-1234",
			[]USBDevice{
				{
					Serial:     "dut-serial-1234",
					DevicePath: "/sys/bus/usb/devices/1-2.3.4/",
					HubPath:    "/sys/bus/usb/devices/1-2.3",
					DeviceType: "cr50",
				},
				{
					Serial:     "servo-serial-1234",
					DevicePath: "/sys/bus/usb/devices/1-2.3.3/",
					HubPath:    "/sys/bus/usb/devices/1-2.3",
					DeviceType: "servo4.1",
				},
			},
			USBDevice{
				Serial:     "servo-serial-1234",
				DevicePath: "/sys/bus/usb/devices/1-2.3.3/",
				HubPath:    "/sys/bus/usb/devices/1-2.3",
				DeviceType: "servo4.1",
			},
			nil,
		},
	}

	for _, tt := range cases {
		expected := tt.output
		actual, err := FindServoFromDUT(tt.inputSerial, tt.inputDeviceList)

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("expected: %v, unexpected diff: %s", expected, diff)
		}

		if !errors.Is(err, tt.err) {
			t.Errorf("expected: %v, got: %v", tt.err, err)
		}
	}
}
