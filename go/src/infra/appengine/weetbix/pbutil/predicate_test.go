// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package pbutil

import (
	"testing"

	atvpb "infra/appengine/weetbix/proto/analyzedtestvariant"
	pb "infra/appengine/weetbix/proto/v1"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestValidateAnalyzedTestVariantPredicate(t *testing.T) {
	Convey(`TestValidateAnalyzedTestVariantPredicate`, t, func() {
		Convey(`Empty`, func() {
			err := ValidateAnalyzedTestVariantPredicate(&atvpb.Predicate{})
			So(err, ShouldBeNil)
		})

		Convey(`TestID`, func() {
			validate := func(TestIdRegexp string) error {
				return ValidateAnalyzedTestVariantPredicate(&atvpb.Predicate{
					TestIdRegexp: TestIdRegexp,
				})
			}

			Convey(`empty`, func() {
				So(validate(""), ShouldBeNil)
			})

			Convey(`valid`, func() {
				So(validate("A.+"), ShouldBeNil)
			})

			Convey(`invalid`, func() {
				So(validate(")"), ShouldErrLike, "test_id_regexp: error parsing regex")
			})
			Convey(`^`, func() {
				So(validate("^a"), ShouldErrLike, "test_id_regexp: must not start with ^")
			})
			Convey(`$`, func() {
				So(validate("a$"), ShouldErrLike, "test_id_regexp: must not end with $")
			})
		})

		Convey(`Test variant`, func() {
			validVariant := Variant("a", "b")
			invalidVariant := Variant("", "")

			validate := func(p *pb.VariantPredicate) error {
				return ValidateAnalyzedTestVariantPredicate(&atvpb.Predicate{
					Variant: p,
				})
			}

			Convey(`Equals`, func() {
				Convey(`Valid`, func() {
					err := validate(&pb.VariantPredicate{
						Predicate: &pb.VariantPredicate_Equals{Equals: validVariant},
					})
					So(err, ShouldBeNil)
				})
				Convey(`Invalid`, func() {
					err := validate(&pb.VariantPredicate{
						Predicate: &pb.VariantPredicate_Equals{Equals: invalidVariant},
					})
					So(err, ShouldErrLike, `variant: equals: "":"": key: unspecified`)
				})
			})

			Convey(`Contains`, func() {
				Convey(`Valid`, func() {
					err := validate(&pb.VariantPredicate{
						Predicate: &pb.VariantPredicate_Contains{Contains: validVariant},
					})
					So(err, ShouldBeNil)
				})
				Convey(`Invalid`, func() {
					err := validate(&pb.VariantPredicate{
						Predicate: &pb.VariantPredicate_Contains{Contains: invalidVariant},
					})
					So(err, ShouldErrLike, `variant: contains: "":"": key: unspecified`)
				})
			})

			Convey(`Unspecified`, func() {
				err := validate(&pb.VariantPredicate{})
				So(err, ShouldErrLike, `variant: unspecified`)
			})
		})

		Convey(`Status`, func() {
			validate := func(s atvpb.Status) error {
				return ValidateAnalyzedTestVariantPredicate(&atvpb.Predicate{
					Status: s,
				})
			}
			Convey(`unspecified`, func() {
				err := validate(atvpb.Status_STATUS_UNSPECIFIED)
				So(err, ShouldBeNil)
			})
			Convey(`invalid`, func() {
				err := validate(atvpb.Status(100))
				So(err, ShouldErrLike, `status: invalid value 100`)
			})
			Convey(`valid`, func() {
				err := validate(atvpb.Status_FLAKY)
				So(err, ShouldBeNil)
			})
		})
	})
}
