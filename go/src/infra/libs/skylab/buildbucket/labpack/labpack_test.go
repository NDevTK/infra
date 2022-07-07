// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package labpack

import (
	"context"
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"

	"infra/libs/skylab/buildbucket"
)

// TestAsMap tests structbuilder-compatibility.
//
// Make sure that we only have keys of a type that structbuilder understands.
//
// We will be more conservative than structbuilder and reject everything that isn't a bool or a string.
//
// Keep the function deterministic by sorting the keys before we check for
// values that have an unsupported type.
func TestAsMap(t *testing.T) {
	t.Parallel()
	zero := Params{}
	zeroMap := zero.AsMap()

	// Keep the function deterministic by sorting the keys before we check for
	// values that have an unsupported type.
	keys := make([]string, 0, len(zeroMap))
	for k := range zeroMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := zeroMap[k]
		switch v := v.(type) {
		case bool, string:
			// do nothing
		default:
			t.Errorf("key %q has value %v with unsupported type %T", k, v, v)
		}
	}
}

type FakeClient struct {
	startID int64
}

func (c *FakeClient) ScheduleLabpackTask(ctx context.Context, params *buildbucket.ScheduleLabpackTaskParams) (string, int64, error) {
	id := c.startID
	c.startID++
	return "", id, nil
}

func (c *FakeClient) BuildURL(buildID int64) string {
	panic("BuildURL should not be called!")
}

// TestScheduleTask tests whether schedule task accepts or rejects its arguments, basically.
func TestScheduleTask(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("test schedule task", t, func() {
		Convey("nil params", func() {
			_, err := ScheduleTask(ctx, &FakeClient{}, CIPDProd, nil)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "schedule task")
		})
		Convey("audit-rpm", func() {
			bbid, err := ScheduleTask(ctx, &FakeClient{}, CIPDProd, &Params{
				BuilderName: "audit-rpm",
			})
			So(err, ShouldBeNil)
			So(bbid, ShouldEqual, 0)
		})
	})
}
