// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	ufspb "infra/unifiedfleet/api/v1/proto"
)

func TestIsLocation(t *testing.T) {
	Convey("test standard chromeos location", t, func() {
		location := []string{
			"chromeos1-row2-rack3-host4",
			"chromeos1-row2-rack3-hostxxx",
			"ROW1-RACK2-HOST3",
			"chromeos6-floor",
			"chromeos6-rack1",
			"chromeos6-storage1",
			"2081-storage1",
			"em25-desk-stagenut",
			"chromeos6-row2-rack23-labstation1",
			"chromeos6-row2-rack23-labstation",
		}
		for _, l := range location {
			So(IsLocation(l), ShouldBeTrue)
		}
	})

	Convey("test invalid chromeos location", t, func() {
		location := "chromeos1-row2-rack3"
		So(IsLocation(location), ShouldBeFalse)
	})
}

func TestUFSLabCoverage(t *testing.T) {
	Convey("test the ufs lab mapping covers all UFS lab enum", t, func() {
		got := make(map[string]bool, len(StrToUFSLab))
		for _, v := range StrToUFSLab {
			got[v] = true
		}
		for l := range ufspb.Lab_value {
			if l == ufspb.Lab_LAB_UNSPECIFIED.String() {
				continue
			}
			_, ok := got[l]
			So(ok, ShouldBeTrue)
		}
	})

	Convey("test the ufs lab mapping doesn't cover any non-UFS lab enum", t, func() {
		for _, v := range StrToUFSLab {
			_, ok := ufspb.Lab_value[v]
			So(ok, ShouldBeTrue)
		}
	})
}

func TestUFSStateCoverage(t *testing.T) {
	Convey("test the ufs state mapping covers all UFS state enum", t, func() {
		got := make(map[string]bool, len(StrToUFSState))
		for _, v := range StrToUFSState {
			got[v] = true
		}
		for l := range ufspb.State_value {
			if l == ufspb.State_STATE_UNSPECIFIED.String() {
				continue
			}
			_, ok := got[l]
			So(ok, ShouldBeTrue)
		}
	})

	Convey("test the ufs state mapping doesn't cover any non-UFS state enum", t, func() {
		for _, v := range StrToUFSState {
			_, ok := ufspb.State_value[v]
			So(ok, ShouldBeTrue)
		}
	})
}

func TestReplaceLabNames(t *testing.T) {
	Convey("TestReplaceLabNames labname error", t, func() {
		filter := "machine=mac-1,mac-2 & lab=XXX & nic=nic-1"
		_, err := ReplaceLabNames(filter)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Invalid lab name XXX for filtering.")
	})

	Convey("TestReplaceLabNames labname - happy path", t, func() {
		filter := "machine=mac-1,mac-2 & lab=atl97,mtv96 & nic=nic-1"
		filter, _ = ReplaceLabNames(filter)
		So(filter, ShouldNotBeNil)
		So(filter, ShouldContainSubstring, "machine=mac-1,mac-2&lab=LAB_DATACENTER_ATL97,LAB_DATACENTER_MTV96&nic=nic-1")
	})
}
