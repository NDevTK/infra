// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTrimByteString(t *testing.T) {
	Convey("Test Trimming", t, func() {
		byteStrings := map[string]string{
			"   5,000 MB   ": "5000MB",
			"6.000 GB   ":    "6000GB",
			"   70 00":       "7000",
			"a\\&*c":         "a\\&*c",
		}
		for k, v := range byteStrings {
			So(TrimByteString(k), ShouldEqual, v)
		}
	})
}

func TestConvertToBytes(t *testing.T) {
	Convey("Test Byte String Conversion", t, func() {
		Convey("Empty String", func() {
			bytes, err := ConvertToBytes("0")
			So(err, ShouldBeNil)
			So(bytes, ShouldBeZeroValue)
		})
		Convey("Valid String with Unit Suffix", func() {
			bytes, err := ConvertToBytes("500KB")
			So(err, ShouldBeNil)
			So(bytes, ShouldEqual, 500000)
		})
		Convey("Valid String No Unit Suffix", func() {
			bytes, err := ConvertToBytes("1000")
			So(err, ShouldBeNil)
			So(bytes, ShouldEqual, 1000)
		})
		Convey("Invalid Numeric", func() {
			bytes, err := ConvertToBytes("E39")
			So(err, ShouldNotBeNil)
			So(bytes, ShouldEqual, 0)
		})
		Convey("Invalid Unit Suffix", func() {
			bytes, err := ConvertToBytes("100Giraffe")
			So(err, ShouldNotBeNil)
			So(bytes, ShouldEqual, 0)
		})
	})
}
