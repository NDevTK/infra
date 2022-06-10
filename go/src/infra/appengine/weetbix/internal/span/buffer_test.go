// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package span

import (
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/golang/protobuf/proto"
	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/protobuf/types/known/timestamppb"

	"infra/appengine/weetbix/pbutil"
	atvpb "infra/appengine/weetbix/proto/analyzedtestvariant"
	pb "infra/appengine/weetbix/proto/v1"
)

func TestTypeConversion(t *testing.T) {
	t.Parallel()

	var b Buffer

	test := func(goValue, spValue interface{}) {
		// ToSpanner
		actualSPValue := ToSpanner(goValue)
		So(actualSPValue, ShouldResemble, spValue)

		// FromSpanner
		row, err := spanner.NewRow([]string{"a"}, []interface{}{actualSPValue})
		So(err, ShouldBeNil)
		goPtr := reflect.New(reflect.TypeOf(goValue))
		err = b.FromSpanner(row, goPtr.Interface())
		So(err, ShouldBeNil)
		So(goPtr.Elem().Interface(), ShouldResemble, goValue)
	}

	Convey(`int64`, t, func() {
		test(int64(42), int64(42))
	})

	Convey(`*timestamppb.Timestamp`, t, func() {
		test(
			&timestamppb.Timestamp{Seconds: 1000, Nanos: 1234},
			spanner.NullTime{Valid: true, Time: time.Unix(1000, 1234).UTC()},
		)
	})

	Convey(`atvpb.Status`, t, func() {
		test(atvpb.Status_STATUS_UNSPECIFIED, int64(0))
	})

	Convey(`pb.BuildStatus`, t, func() {
		test(pb.BuildStatus_BUILD_STATUS_SUCCESS, int64(1))
	})
	Convey(`[]pb.ExonerationReason`, t, func() {
		test(
			[]pb.ExonerationReason{
				pb.ExonerationReason_OCCURS_ON_MAINLINE,
				pb.ExonerationReason_NOT_CRITICAL,
			},
			[]int64{int64(1), int64(3)},
		)
		test(
			[]pb.ExonerationReason{},
			[]int64{},
		)
		test(
			[]pb.ExonerationReason(nil),
			[]int64(nil),
		)
	})
	Convey(`pb.TestResultStatus`, t, func() {
		test(pb.TestResultStatus_PASS, int64(1))
	})

	Convey(`*pb.Variant`, t, func() {
		Convey(`Works`, func() {
			test(
				pbutil.Variant("a", "1", "b", "2"),
				[]string{"a:1", "b:2"},
			)
		})
		Convey(`Empty`, func() {
			test(
				(*pb.Variant)(nil),
				[]string{},
			)
		})
	})

	Convey(`[]*pb.StringPair`, t, func() {
		test(
			pbutil.StringPairs("a", "1", "b", "2"),
			[]string{"a:1", "b:2"},
		)
	})

	Convey(`Compressed`, t, func() {
		Convey(`Empty`, func() {
			test(Compressed(nil), []byte(nil))
		})
		Convey(`non-Empty`, func() {
			test(
				Compressed("aaaaaaaaaaaaaaaaaaaa"),
				[]byte{122, 116, 100, 10, 40, 181, 47, 253, 4, 0, 69, 0, 0, 8, 97, 1, 84, 1, 2, 16, 4, 247, 175, 71, 227})
		})
	})

	Convey(`Map`, t, func() {
		var varIntA, varIntB int64
		var varState atvpb.Status

		row, err := spanner.NewRow([]string{"a", "b", "c"}, []interface{}{int64(42), int64(56), int64(0)})
		So(err, ShouldBeNil)
		err = b.FromSpanner(row, &varIntA, &varIntB, &varState)
		So(err, ShouldBeNil)
		So(varIntA, ShouldEqual, 42)
		So(varIntB, ShouldEqual, 56)
		So(varState, ShouldEqual, atvpb.Status_STATUS_UNSPECIFIED)

		// ToSpanner
		spValues := ToSpannerMap(map[string]interface{}{
			"a": varIntA,
			"b": varIntB,
			"c": varState,
		})
		So(spValues, ShouldResemble, map[string]interface{}{"a": int64(42), "b": int64(56), "c": int64(0)})
	})

	Convey(`proto.Message`, t, func() {
		msg := &atvpb.FlakeStatistics{
			FlakyVerdictRate: 0.5,
		}
		expected, err := proto.Marshal(msg)
		So(err, ShouldBeNil)
		So(ToSpanner(msg), ShouldResemble, expected)

		row, err := spanner.NewRow([]string{"a"}, []interface{}{expected})
		So(err, ShouldBeNil)

		Convey(`success`, func() {
			expectedPtr := &atvpb.FlakeStatistics{}
			err = b.FromSpanner(row, expectedPtr)
			So(err, ShouldBeNil)
			So(expectedPtr, ShouldResembleProto, msg)
		})

		Convey(`Passing nil pointer to fromSpanner`, func() {
			var expectedPtr *atvpb.FlakeStatistics
			err = b.FromSpanner(row, expectedPtr)
			So(err, ShouldErrLike, "nil pointer encountered")
		})
	})

	Convey(`[]atvpb.Status`, t, func() {
		test(
			[]atvpb.Status{
				atvpb.Status_FLAKY,
				atvpb.Status_STATUS_UNSPECIFIED,
			},
			[]int64{int64(10), int64(0)},
		)
	})
}
