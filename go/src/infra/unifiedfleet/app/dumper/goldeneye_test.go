// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"bufio"
	"context"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseGoldenEyeJsonData(t *testing.T) {
	t.Parallel()

	Convey("Parse Data", t, func() {
		Convey("happy path", func() {
			file, err := os.Open("test.json")
			So(err, ShouldBeNil)
			reader := bufio.NewReader(file)

			devices, err := parseGoldenEyeData(context.Background(), reader)
			So(err, ShouldEqual, nil)
			So(devices.Devices, ShouldNotBeNil)
		})
		Convey("parse for non existent file", func() {
			file, err := os.Open("test2.json")
			So(err, ShouldNotBeNil)
			reader := bufio.NewReader(file)

			devices, err := parseGoldenEyeData(context.Background(), reader)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unmarshal chunk failed while reading golden eye data for devices: invalid argument")
			So(devices, ShouldBeNil)
		})
	})
}
