// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This package contains constants for recoverylib, including task names for example.
// For more information, see b:208688399.
package buildbucket

import (
	"errors"
	"fmt"
	"strings"
)

// TaskName describes which flow/plans will be involved in the process.
type TaskName string

const (
	InvalidTaskName TaskName = ""
	// Audit is used to run audit task of RPM.
	AuditRPM TaskName = "audit_rpm"
	// Audit is used to run audit task of internal storage.
	AuditStorage TaskName = "audit_storage"
	// Audit is used to run audit task of USB drive.
	AuditUSB TaskName = "audit_usb"
	// Task used to run auto recovery/repair flow in the lab.
	Recovery TaskName = "recovery"
	// Task used to run deep repair flow in the lab.
	DeepRecovery TaskName = "deep_recovery"
	// Task used to prepare device to be used in the lab.
	Deploy TaskName = "deploy"
	// Task used to execute custom plans.
	// Configuration has to be provided by the user.
	Custom TaskName = "custom"
)

// String returns the name of the task as an argument to the labpack command-line tool.
func (tn TaskName) String() string {
	return string(tn)
}

func (tn TaskName) BuilderName() string {
	if name, ok := builderNameMap[tn]; ok {
		return name
	}
	return string(tn)
}

var builderNameMap = map[TaskName]string{
	AuditRPM:     "audit-rpm",
	AuditStorage: "audit-storage",
	AuditUSB:     "audit-servo-usb-key",
	Recovery:     "repair",
	DeepRecovery: "deep-repair",
}

// NormalizeTaskName takes a task name from anywhere and normalizes it.
// This is a necessary first step towards consolidating our notion of task names.
//
// Names are taken from here and https://chromium.googlesource.com/infra/infra/+/refs/heads/main/go/src/infra/appengine/crosskylabadmin/internal/app/frontend/tracker.go .
func NormalizeTaskName(name string) (TaskName, error) {
	switch strings.ToLower(name) {
	case "verify-servo-usb-drive", "usb-drive", "audit-usb", "audit_usb":
		return AuditUSB, nil
	case "verify-dut-storage", "storage", "audit-storage", "audit_storage":
		return AuditStorage, nil
	case "verify-rpm-config", "rpm config", "audit-rpm", "audit_rpm":
		return AuditRPM, nil
	case "repair", "recovery":
		return Recovery, nil
	case "deep-repair", "deep_repair":
		return DeepRecovery, nil
	case "deploy":
		return Deploy, nil
	case "custom":
		return Custom, nil
	}
	return InvalidTaskName, fmt.Errorf("normalize task name: unrecognized task name %q", name)
}

// ValidateTaskName checks whether a task name is valid
func ValidateTaskName(tn TaskName) error {
	if tn == "" {
		return errors.New("validate task name: task name cannot be empty")
	}
	switch tn {
	case AuditRPM:
	case AuditStorage:
	case AuditUSB:
	case Recovery:
	case DeepRecovery:
	case Deploy:
	case Custom:
	default:
		return fmt.Errorf("validate task name: %q is not a valid task name", tn)
	}
	return nil
}
