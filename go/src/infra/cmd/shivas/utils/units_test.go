// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestVerifyByteUnit(t *testing.T) {
	Convey("Test Valid Byte Unit String", t, func() {
		byteUnits := []string{
			"B",
			"KiB",
		}
		for _, bu := range byteUnits {
			So(VerifyByteUnit(bu), ShouldBeNil)
		}
	})
	Convey("Test Invalid Byte Unit String", t, func() {
		byteUnits := []string{
			"",
			"mb",
			"chrome",
		}
		for _, bu := range byteUnits {
			So(VerifyByteUnit(bu), ShouldNotBeNil)
		}
	})
}

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

func TestGetMultipleForByteUnit(t *testing.T) {
	Convey("Test Fetching Multiple for Valid Byte Unit", t, func() {
		byteUnits := map[string]int64{
			"B":   1,
			"KiB": 1024,
			"MiB": 1024 * 1024,
		}
		for k, v := range byteUnits {
			multiple, err := GetMultipleForByteUnit(k)
			So(err, ShouldBeNil)
			So(multiple, ShouldEqual, v)
		}
	})
	Convey("Test Fetching Multiple for Invalid Byte Unit", t, func() {
		byteUnits := []string{
			"",
			"mb",
			"chrome",
		}
		for _, byteUnit := range byteUnits {
			multiple, err := GetMultipleForByteUnit(byteUnit)
			So(err, ShouldNotBeNil)
			So(multiple, ShouldEqual, 0)
		}
	})
}
