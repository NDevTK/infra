// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package kbqpb

import (
	"testing"

	"cloud.google.com/go/bigquery"
	. "github.com/smartystreets/goconvey/convey"
)

// TestObservationIsValueSaver tests that a pointer to an observation implements the ValueSaver interface.
// This test is trivial at runtime but nontrivial at compile time.
func TestObservationIsValueSaver(t *testing.T) {
	t.Parallel()
	var _ bigquery.ValueSaver = &Observation{}
}

// TestObservationSave tests saving an observation. New subtests should be added here for new fields
// to make sure that they are exported correctly.
func TestObservationSave(t *testing.T) {
	t.Parallel()
	Convey("test observation save", t, func() {
		Convey("action name", func() {
			m, _, err := (&Observation{Name: "aaaaa", ActionName: "hi"}).Save()
			So(err, ShouldBeNil)
			So(m["action_name"], ShouldEqual, "hi")
		})
	})
}
