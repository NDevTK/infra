// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package buildbucket

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestValidateTaskName tests that task names are validated correctly.
func TestValidateTaskName(t *testing.T) {
	t.Parallel()
	Convey("validate", t, func() {
		So(ValidateTaskName(""), ShouldNotBeNil)
		So(ValidateTaskName("audit_rpm"), ShouldBeNil)
		So(ValidateTaskName("deep_recovery"), ShouldBeNil)
		So(ValidateTaskName("audit____"), ShouldNotBeNil)
	})
}

var TaskNameToBuilderPerVersionCases = []struct {
	want     string
	taskName TaskName
	version  CIPDVersion
}{
	{"audit-rpm", AuditRPM, CIPDProd},
	{"audit-rpm-latest", AuditRPM, CIPDLatest},
	{"audit-storage", AuditStorage, CIPDProd},
	{"audit-storage-latest", AuditStorage, CIPDLatest},
	{"audit-servo-usb-key", AuditUSB, CIPDProd},
	{"audit-servo-usb-key-latest", AuditUSB, CIPDLatest},
	{"repair", Recovery, CIPDProd},
	{"repair-latest", Recovery, CIPDLatest},
	{"repair", DeepRecovery, CIPDProd},
	{"repair-latest", DeepRecovery, CIPDLatest},
	{"deploy", Deploy, CIPDProd},
	{"deploy-latest", Deploy, CIPDLatest},
	{"custom", Custom, CIPDProd},
	{"custom-latest", Custom, CIPDLatest},
	{"custom", InvalidTaskName, CIPDProd},
	{"custom-latest", InvalidTaskName, CIPDLatest},
}

func TestTaskNameToBuilderPerVersion(t *testing.T) {
	for i, c := range TaskNameToBuilderPerVersionCases {
		name := fmt.Sprintf("case: %d", i)
		c := c
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := TaskNameToBuilderNamePerVersion(c.taskName, c.version)
			if got != c.want {
				t.Errorf("received wrong value: wanted %q but got %q", c.want, got)
			}
		})
	}
}
