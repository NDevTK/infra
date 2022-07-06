// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package kbqpb

import (
	"testing"

	"cloud.google.com/go/bigquery"
	. "github.com/smartystreets/goconvey/convey"
)

// TestActionIsValueSaver tests that a pointer to an action implements the ValueSaver interface.
// This test is trivial at runtime but nontrivial at compile time.
func TestActionIsValueSaver(t *testing.T) {
	t.Parallel()
	var _ bigquery.ValueSaver = &Action{}
}

// TestActionSave tests saving an action. New subtests should be added here for new fields
// to make sure that they are exported correctly.
func TestActionSave(t *testing.T) {
	t.Parallel()
	Convey("test action save", t, func() {
		Convey("model", func() {
			m, _, err := (&Action{Name: "aaaaa", Model: "hi"}).Save()
			So(err, ShouldBeNil)
			So(m["model"], ShouldEqual, "hi")
		})
		Convey("board", func() {
			m, _, err := (&Action{Name: "aaaaa", Board: "hi"}).Save()
			So(err, ShouldBeNil)
			So(m["board"], ShouldEqual, "hi")
		})
	})
}
