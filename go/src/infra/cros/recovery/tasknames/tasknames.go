// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This package contains constants for recoverylib, including task names for example.
// For more information, see b:208688399.
package tasknames

import (
	"errors"
	"fmt"
)

// TaskName describes which flow/plans will be involved in the process.
type TaskName = string

const (
	// Audit is used to run audit task of RPM.
	AuditRPM TaskName = "audit_rpm"
	// Audit is used to run audit task of internal storage.
	AuditStorage TaskName = "audit_storage"
	// Audit is used to run audit task of USB drive.
	AuditUSB TaskName = "audit_usb"
	// Task used to run auto recovery/repair flow in the lab.
	Recovery TaskName = "recovery"
	// Task used to prepare device to be used in the lab.
	Deploy TaskName = "deploy"
	// Task used to execute custom plans.
	// Configuration has to be provided by the user.
	Custom TaskName = "custom"
)

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
	case Deploy:
	case Custom:
	default:
		return fmt.Errorf("validate task name: %q is not a valid task name", tn)
	}
	return nil
}
