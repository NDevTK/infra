// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package launcher

import (
	"testing"

	ds "github.com/luci/gae/service/datastore"
	tq "github.com/luci/gae/service/taskqueue"

	. "github.com/smartystreets/goconvey/convey"

	"infra/tricium/appengine/common"
	"infra/tricium/appengine/common/pipeline"
	trit "infra/tricium/appengine/common/testing"
)

func TestAnalyzeRequest(t *testing.T) {
	Convey("Test Environment", t, func() {
		tt := &trit.Testing{}
		ctx := tt.Context()

		project := "test-project"
		gitref := "ref/test"
		paths := []string{
			"README.md",
			"README2.md",
		}
		lr := &pipeline.LaunchRequest{
			RunID:   123456789,
			Project: project,
			GitRef:  gitref,
			Path:    paths,
		}

		Convey("Launch request", func() {
			err := launch(ctx, lr)
			So(err, ShouldBeNil)

			Convey("Enqueues track request", func() {
				So(len(tq.GetTestable(ctx).GetScheduledTasks()[common.TrackerQueue]), ShouldEqual, 1)
			})

			Convey("Stores workflow config", func() {
				wf := &common.Entity{
					ID:   lr.RunID,
					Kind: "Workflow",
				}
				err := ds.Get(ctx, wf)
				So(err, ShouldBeNil)
			})

			// TODO(emso): Sends driver requests for root workers.
			// TODO(emso): Lists root workers from workflow config.
		})
	})
}
