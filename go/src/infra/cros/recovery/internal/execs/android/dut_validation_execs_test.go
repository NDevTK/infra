// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package android

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/tlw"
)

func TestHasDutBoardExec(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("hasDutBoardExec", t, func() {
		Convey("Attached DUT board is present - no error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Android: &tlw.Android{
							Board:              "board",
							Model:              "model",
							SerialNumber:       "serialNumber",
							AssociatedHostname: "associatedHostname",
						},
					},
				}, "some name", nil, 0)
			So(hasDutBoardExec(ctx, info), ShouldBeNil)
		})
		Convey("Missing attached DUT board - returns error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Android: &tlw.Android{
							Model:              "model",
							SerialNumber:       "serialNumber",
							AssociatedHostname: "associatedHostname",
						},
					},
				}, "", nil, 0)
			So(hasDutBoardExec(ctx, info), ShouldNotBeNil)
		})
		Convey("ChromeOs DUT  with board - returns error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Chromeos: &tlw.ChromeOS{
							Board:        "board",
							Model:        "model",
							SerialNumber: "serialNumber",
						},
					},
				}, "", nil, 0)
			So(hasDutBoardExec(ctx, info), ShouldNotBeNil)
		})
	})
}

func TestHasDutModelExec(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("hasDutModelExec", t, func() {
		Convey("Attached DUT model is present - no error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Android: &tlw.Android{
							Board:              "board",
							Model:              "model",
							SerialNumber:       "serialNumber",
							AssociatedHostname: "associatedHostname",
						},
					},
				}, "", nil, 0,
			)
			So(hasDutModelExec(ctx, info), ShouldBeNil)
		})
		Convey("Missing attached DUT model - returns error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Android: &tlw.Android{
							Board:              "board",
							SerialNumber:       "serialNumber",
							AssociatedHostname: "associatedHostname",
						},
					},
				}, "name", nil, 0)
			So(hasDutModelExec(ctx, info), ShouldNotBeNil)
		})
		Convey("ChromeOs DUT with model - returns error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Chromeos: &tlw.ChromeOS{
							Board:        "board",
							Model:        "model",
							SerialNumber: "serialNumber",
						},
					},
				}, "", nil, 0)

			So(hasDutModelExec(ctx, info), ShouldNotBeNil)
		})
	})
}

func TestHasDutSerialNumberExec(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("hasDutSerialNumberExec", t, func() {
		Convey("Attached DUT serial number is present - no error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Android: &tlw.Android{
							Board:              "board",
							Model:              "model",
							SerialNumber:       "serialNumber",
							AssociatedHostname: "associatedHostname",
						},
					},
				}, "", nil, 0)
			So(hasDutSerialNumberExec(ctx, info), ShouldBeNil)
		})
		Convey("Missing attached DUT serial number - returns error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Android: &tlw.Android{
							Board:              "board",
							Model:              "model",
							AssociatedHostname: "associatedHostname",
						},
					},
				}, "", nil, 0)
			So(hasDutSerialNumberExec(ctx, info), ShouldNotBeNil)
		})
		Convey("ChromeOs DUT with serial number - returns error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Chromeos: &tlw.ChromeOS{
							Board:        "board",
							Model:        "model",
							SerialNumber: "serialNumber",
						},
					},
				}, "", nil, 0)
			So(hasDutSerialNumberExec(ctx, info), ShouldNotBeNil)
		})
	})
}

func TestHasDutAssociatedHostExec(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("hasDutAssociatedHostExec", t, func() {
		Convey("Attached DUT associated hostname is present - no error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Android: &tlw.Android{
							Board:              "board",
							Model:              "model",
							SerialNumber:       "serialNumber",
							AssociatedHostname: "associatedHostname",
						},
					},
				}, "", nil, 0)
			So(hasDutAssociatedHostExec(ctx, info), ShouldBeNil)
		})
		Convey("Missing attached DUT associated hostname - returns error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Android: &tlw.Android{
							Board:        "board",
							Model:        "model",
							SerialNumber: "serialNumber",
						},
					},
				}, "", nil, 0)
			So(hasDutAssociatedHostExec(ctx, info), ShouldNotBeNil)
		})
		Convey("ChromeOs DUT - returns error", func() {
			info := execs.NewExecInfo(
				&execs.RunArgs{
					DUT: &tlw.Dut{
						Chromeos: &tlw.ChromeOS{
							Board:        "board",
							Model:        "model",
							SerialNumber: "serialNumber",
						},
					},
				}, "", nil, 0)
			So(hasDutAssociatedHostExec(ctx, info), ShouldNotBeNil)
		})
	})
}
