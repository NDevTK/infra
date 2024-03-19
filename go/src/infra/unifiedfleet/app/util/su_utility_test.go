// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSchedulingUnitDutState(t *testing.T) {
	Convey("Test when all child DUTs in ready.", t, func() {
		s := []string{"ready", "ready", "ready", "ready", "ready"}
		So(SchedulingUnitDutState(s), ShouldEqual, "ready")
	})

	Convey("Test when where one child DUT in needs_repair.", t, func() {
		s := []string{"ready", "ready", "ready", "ready", "needs_repair"}
		So(SchedulingUnitDutState(s), ShouldEqual, "needs_repair")
	})

	Convey("Test when where one child DUT in repair_failed.", t, func() {
		s := []string{"ready", "ready", "ready", "needs_repair", "repair_failed"}
		So(SchedulingUnitDutState(s), ShouldEqual, "repair_failed")
	})

	Convey("Test when where one child DUT in needs_manual_repair.", t, func() {
		s := []string{"ready", "ready", "needs_manual_repair", "needs_repair", "repair_failed"}
		So(SchedulingUnitDutState(s), ShouldEqual, "needs_manual_repair")
	})

	Convey("Test when where one child DUT in needs_replacement.", t, func() {
		s := []string{"ready", "needs_deploy", "needs_replacement", "needs_repair", "needs_manual_repair"}
		So(SchedulingUnitDutState(s), ShouldEqual, "needs_replacement")
	})

	Convey("Test when where one child DUT in needs_deploy.", t, func() {
		s := []string{"ready", "ready", "needs_deploy", "needs_manual_repair", "repair_failed"}
		So(SchedulingUnitDutState(s), ShouldEqual, "needs_deploy")
	})

	Convey("Test when where one child DUT in reserved.", t, func() {
		s := []string{"ready", "reserved", "needs_deploy", "needs_repair", "needs_replacement"}
		So(SchedulingUnitDutState(s), ShouldEqual, "reserved")
	})

	Convey("Test when input is an empty slice", t, func() {
		var s []string
		So(SchedulingUnitDutState(s), ShouldEqual, "unknown")
	})
}
