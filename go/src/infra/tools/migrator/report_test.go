// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package migrator

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/common/data/stringset"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/config"
)

func TestReportID(t *testing.T) {
	t.Parallel()

	Convey(`ReportID`, t, func() {
		Convey(`ConfigSet`, func() {
			So(ReportID{Project: "foo"}.ConfigSet(), ShouldResemble,
				config.Set("projects/foo"))
			So(ReportID{Project: "foo", ConfigFile: "irrelevant"}.ConfigSet(), ShouldResemble,
				config.Set("projects/foo"))
		})

		Convey(`String`, func() {
			So(ReportID{Checkout: "checkout"}.String(), ShouldResemble, "checkout")
			So(ReportID{Checkout: "checkout", Project: "foo"}.String(), ShouldResemble, "checkout|foo")
			So(ReportID{Checkout: "checkout", Project: "foo", ConfigFile: "file"}.String(), ShouldResemble, "checkout|foo|file")
		})
	})
}

func TestReport(t *testing.T) {
	t.Parallel()

	Convey(`Report`, t, func() {
		r := &Report{
			ReportID: ReportID{"checkout", "proj-foo", "config.file"},
			Tag:      "SOME_TAG",
			Problem:  "This is a problem.",
			Metadata: map[string]stringset.Set{
				"meta": stringset.NewFromSlice("value"),
			},
		}

		Convey(`Clone`, func() {
			ptr := func(a any) uintptr {
				return reflect.ValueOf(a).Pointer()
			}

			newR := r.Clone()
			So(r.ReportID, ShouldResemble, newR.ReportID)
			So(r.Tag, ShouldResemble, newR.Tag)
			So(r.Problem, ShouldResemble, newR.Problem)
			So(ptr(r.Metadata), ShouldNotEqual, ptr(newR.Metadata))                 // different maps
			So(ptr(r.Metadata["meta"]), ShouldNotEqual, ptr(newR.Metadata["meta"])) // different sets
			So(r.Metadata["meta"].ToSlice(), ShouldResemble, newR.Metadata["meta"].ToSlice())
		})

		Convey(`ToCSVRow`, func() {
			So(r.ToCSVRow(), ShouldResemble, []string{
				"checkout", "proj-foo", "config.file", "SOME_TAG", "This is a problem.", "false",
				"meta:value",
			})

			r.Actionable = true
			So(r.ToCSVRow(), ShouldResemble, []string{
				"checkout", "proj-foo", "config.file", "SOME_TAG", "This is a problem.", "true",
				"meta:value",
			})
		})

		Convey(`NewReportFromCSVRow`, func() {
			Convey(`Good`, func() {
				report, err := NewReportFromCSVRow([]string{
					"checkout", "proj-foo", "config.file", "SOME_TAG", "This is a problem.",
					"true", "meta:value", "meta:other_value", "other_meta:1",
				})
				So(err, ShouldBeNil)
				So(report.ReportID, ShouldResemble, ReportID{"checkout", "proj-foo", "config.file"})
				So(report.Tag, ShouldResemble, "SOME_TAG")
				So(report.Problem, ShouldResemble, "This is a problem.")
				So(report.Actionable, ShouldBeTrue)
				So(report.Metadata, ShouldResemble, map[string]stringset.Set{
					"meta":       stringset.NewFromSlice("value", "other_value"),
					"other_meta": stringset.NewFromSlice("1"),
				})
			})

			Convey(`Bad`, func() {
				Convey(`no Checkout`, func() {
					_, err := NewReportFromCSVRow(nil)
					So(err, ShouldErrLike, "Checkout field")

					_, err = NewReportFromCSVRow([]string{""})
					So(err, ShouldErrLike, "Checkout field")
				})

				Convey(`no Project`, func() {
					_, err := NewReportFromCSVRow([]string{"checkout"})
					So(err, ShouldErrLike, "Project field")

					_, err = NewReportFromCSVRow([]string{"checkout", ""})
					So(err, ShouldErrLike, "Project field")
				})

				Convey(`no ConfigFile`, func() {
					_, err := NewReportFromCSVRow([]string{"checkout", "proj-foo"})
					So(err, ShouldErrLike, "ConfigFile field")

					_, err = NewReportFromCSVRow([]string{"checkout", "proj-foo", ""})
					So(err, ShouldErrLike, "Tag field")
				})

				Convey(`no Tag`, func() {
					_, err := NewReportFromCSVRow([]string{"checkout", "proj-foo", ""})
					So(err, ShouldErrLike, "Tag field")

					_, err = NewReportFromCSVRow([]string{"checkout", "proj-foo", "", ""})
					So(err, ShouldErrLike, "Tag field")
				})

				Convey(`no Problem`, func() {
					_, err := NewReportFromCSVRow([]string{"checkout", "proj-foo", "", "TAG"})
					So(err, ShouldErrLike, "Problem field")

					_, err = NewReportFromCSVRow([]string{"checkout", "proj-foo", "", "TAG", ""})
					So(err, ShouldErrLike, "Actionable field")
				})

				Convey(`no Actionable`, func() {
					_, err := NewReportFromCSVRow([]string{"checkout", "proj-foo", "", "TAG", ""})
					So(err, ShouldErrLike, "Actionable field")

					_, err = NewReportFromCSVRow([]string{"checkout", "proj-foo", "", "TAG", "", "true"})
					So(err, ShouldBeNil)
				})

				Convey(`bad metadata`, func() {
					_, err := NewReportFromCSVRow([]string{"checkout", "proj-foo", "", "TAG", "", "true", "bad"})
					So(err, ShouldErrLike, "Malformed metadata")

					_, err = NewReportFromCSVRow([]string{"checkout", "proj-foo", "", "TAG", "", "true", "ok:value"})
					So(err, ShouldBeNil)
				})
			})
		})
	})
}
