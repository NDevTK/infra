// Copyright (c) 2017 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package system

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRemoveDiskDevices(t *testing.T) {
	tests := []struct {
		Names []string
		Want  []string
	}{
		{[]string{"sda", "sda1"}, []string{"sda1"}},
		{[]string{"sda1", "sda"}, []string{"sda1"}},
		{[]string{"sda", "sdb1"}, []string{"sda", "sdb1"}},
	}

	for i, test := range tests {
		Convey(fmt.Sprintf("%d. %s", i, test.Names), t, func() {
			So(removeDiskDevices(test.Names), ShouldResemble, test.Want)
		})
	}
}

func TestMountpointsAreIgnored(t *testing.T) {
	t.Parallel()

	Convey("Docker mountpoints are ignored", t, func() {
		So(shouldIgnoreMountpoint("/var/lib/docker/aufs"), ShouldBeTrue)
	})
}
