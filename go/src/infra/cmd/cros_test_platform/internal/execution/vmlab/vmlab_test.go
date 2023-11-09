// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmlab

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConvertBuilderName(t *testing.T) {
	Convey("success", t, func() {
		want := "test_runner_gce"
		got := ConvertBuilderName("test_runner")
		So(got, ShouldEqual, want)
	})

	Convey("success non prod", t, func() {
		want := "test_runner_gce-staging"
		got := ConvertBuilderName("test_runner-staging")
		So(got, ShouldEqual, want)
	})

	Convey("ignore already converted", t, func() {
		want := "test_runner_gce"
		got := ConvertBuilderName("test_runner_gce")
		So(got, ShouldEqual, want)
	})

	Convey("ignore unknown", t, func() {
		want := "trv3"
		got := ConvertBuilderName("trv3")
		So(got, ShouldEqual, want)
	})

	Convey("ignore empty", t, func() {
		want := ""
		got := ConvertBuilderName("")
		So(got, ShouldEqual, want)
	})
}

func TestEligible(t *testing.T) {
	var boards = []string{"betty", "reven-vmtest", "amd64-generic"}
	for _, board := range boards {
		Convey("experiment not enabled", t, func() {
			got := eligible(board, []string{"exp1", "exp2"})
			So(got, ShouldBeFalse)
		})

		Convey("experiment enabled", t, func() {
			got := eligible(board, []string{"exp1", "chromeos.cros_infra_config.vmlab.launch", "exp2"})
			So(got, ShouldBeTrue)
		})
	}
	Convey("unsupported board", t, func() {
		got := eligible("anotherboard", []string{"chromeos.cros_infra_config.vmlab.launch"})
		So(got, ShouldBeFalse)
	})

	Convey("unsupported board and no experiment", t, func() {
		got := eligible("anotherboard", nil)
		So(got, ShouldBeFalse)
	})
}
