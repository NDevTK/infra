// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetHiveForDut(t *testing.T) {
	Convey("Test GetHiveForDut", t, func() {
		So(GetHiveForDut("satlab-xx12sha23-chromeos1-row2-rack3-host4", ""), ShouldEqual, "satlab-xx12sha23")
		So(GetHiveForDut("satlab-12:sha45-em25-desk-noogler", ""), ShouldEqual, "satlab-12:sha45")
		So(GetHiveForDut("satlab-abc123-host1", "satlab-nothash"), ShouldEqual, "satlab-nothash")
		So(GetHiveForDut("chromeos1-row2-rack3-host4", ""), ShouldEqual, "")
		So(GetHiveForDut("cros-mtv1950-144-rack204-host1", ""), ShouldEqual, gtransitHive)
		So(GetHiveForDut("cros-mtv1950-144-rack204-host2", ""), ShouldEqual, gtransitHive)
		So(GetHiveForDut("cros-mtv1950-144-rack204-labstation1", ""), ShouldEqual, gtransitHive)
		So(GetHiveForDut("chromeos8-foo", ""), ShouldEqual, sfo36OSHive)
		So(GetHiveForDut("chrome-chromeos8-foo", ""), ShouldEqual, sfo36OSHive)
		So(GetHiveForDut(ChromiumNamePrefix, ""), ShouldEqual, chromiumHive)
		So(GetHiveForDut("chromium-bar", ""), ShouldEqual, chromiumHive)
		So(GetHiveForDut("chrome-perf-chromeos8-host2", ""), ShouldEqual, chromePerfHive)
		So(GetHiveForDut("cri12-host8", ""), ShouldEqual, iad65OSHive)
	})
}

func TestAppendUnique(t *testing.T) {
	Convey("Test AppendUnique", t, func() {
		So(AppendUniqueStrings([]string{"eeny", "meeny", "miny", "moe"}, "catch", "a", "tiger", "by", "the", "toe"), ShouldHaveLength, 10)
		So(AppendUniqueStrings([]string{"london", "bridge", "is", "falling", "down"}, "falling", "down", "falling", "down"), ShouldHaveLength, 5)
		So(AppendUniqueStrings([]string{}, "twinkle", "twinkle", "little", "star"), ShouldHaveLength, 3)
		So(AppendUniqueStrings([]string{"humpty", "dumpty", "sat", "on", "a", "wall"}), ShouldHaveLength, 6)
		So(AppendUniqueStrings([]string{"row", "row", "row", "your", "boat"}), ShouldHaveLength, 3)
	})
}

func TestIsSFPZone(t *testing.T) {
	t.Parallel()

	Convey("Testing IsSFPZone", t, func() {
		So(IsSFPZone("ZONE_SFP_TEST"), ShouldBeTrue)
		So(IsSFPZone("ZONE_SFP"), ShouldBeFalse)
		So(IsSFPZone("FAKE_SFP_TEST"), ShouldBeFalse)
		So(IsSFPZone("ZONE_OTHER_TEST"), ShouldBeFalse)
	})
}
