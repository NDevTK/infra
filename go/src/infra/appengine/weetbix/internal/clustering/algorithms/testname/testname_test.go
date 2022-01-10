// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testname

import (
	"testing"

	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/clustering/rules/lang"
	"infra/appengine/weetbix/internal/config/compiledcfg"
	configpb "infra/appengine/weetbix/internal/config/proto"
	pb "infra/appengine/weetbix/proto/v1"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAlgorithm(t *testing.T) {
	rules := []*configpb.TestNameClusteringRule{
		{
			Name:         "Blink Web Tests",
			Pattern:      `^ninja://:blink_web_tests/(virtual/[^/]+/)?(?P<testname>([^/]+/)+[^/]+\.[a-zA-Z]+).*$`,
			LikeTemplate: "ninja://:blink\\_web\\_tests/%${testname}%",
		},
	}
	cfgpb := &configpb.ProjectConfig{
		Clustering: &configpb.Clustering{
			TestNameRules: rules,
		},
	}

	Convey(`Cluster`, t, func() {
		a := &Algorithm{}
		cfg, err := compiledcfg.NewConfig(cfgpb)
		So(err, ShouldBeNil)

		Convey(`ID of appropriate length`, func() {
			id := a.Cluster(cfg, &clustering.Failure{
				TestID: "ninja://test_name",
			})
			// IDs may be 16 bytes at most.
			So(len(id), ShouldBeGreaterThan, 0)
			So(len(id), ShouldBeLessThanOrEqualTo, clustering.MaxClusterIDBytes)
		})
		Convey(`Same ID for same test name`, func() {
			Convey(`No matching rules`, func() {
				id1 := a.Cluster(cfg, &clustering.Failure{
					TestID: "ninja://test_name_one/",
					Reason: &pb.FailureReason{PrimaryErrorMessage: "A"},
				})
				id2 := a.Cluster(cfg, &clustering.Failure{
					TestID: "ninja://test_name_one/",
					Reason: &pb.FailureReason{PrimaryErrorMessage: "B"},
				})
				So(id2, ShouldResemble, id1)
			})
			Convey(`Matching rules`, func() {
				id1 := a.Cluster(cfg, &clustering.Failure{
					TestID: "ninja://:blink_web_tests/virtual/abc/folder/test-name.html",
					Reason: &pb.FailureReason{PrimaryErrorMessage: "A"},
				})
				id2 := a.Cluster(cfg, &clustering.Failure{
					TestID: "ninja://:blink_web_tests/folder/test-name.html?param=2",
					Reason: &pb.FailureReason{PrimaryErrorMessage: "B"},
				})
				So(id2, ShouldResemble, id1)
			})
		})
		Convey(`Different ID for different clusters`, func() {
			Convey(`No matching rules`, func() {
				id1 := a.Cluster(cfg, &clustering.Failure{
					TestID: "ninja://test_name_one/",
				})
				id2 := a.Cluster(cfg, &clustering.Failure{
					TestID: "ninja://test_name_two/",
				})
				So(id2, ShouldNotResemble, id1)
			})
			Convey(`Matching rules`, func() {
				id1 := a.Cluster(cfg, &clustering.Failure{
					TestID: "ninja://:blink_web_tests/virtual/abc/folder/test-name-a.html",
					Reason: &pb.FailureReason{PrimaryErrorMessage: "A"},
				})
				id2 := a.Cluster(cfg, &clustering.Failure{
					TestID: "ninja://:blink_web_tests/folder/test-name-b.html?param=2",
					Reason: &pb.FailureReason{PrimaryErrorMessage: "B"},
				})
				So(id2, ShouldNotResemble, id1)
			})
		})
	})
	Convey(`Failure Association Rule`, t, func() {
		a := &Algorithm{}
		cfg, err := compiledcfg.NewConfig(cfgpb)
		So(err, ShouldBeNil)

		test := func(failure *clustering.Failure, expectedRule string) {
			rule := a.FailureAssociationRule(cfg, failure)
			So(rule, ShouldEqual, expectedRule)

			// Test the rule is valid syntax and matches at least the example failure.
			expr, err := lang.Parse(rule)
			So(err, ShouldBeNil)
			So(expr.Evaluate(failure), ShouldBeTrue)
		}
		Convey(`No matching rules`, func() {
			failure := &clustering.Failure{
				TestID: "ninja://test_name_one/",
			}
			test(failure, `test = "ninja://test_name_one/"`)
		})
		Convey(`Matching rule`, func() {
			failure := &clustering.Failure{
				TestID: "ninja://:blink_web_tests/virtual/dark-color-scheme/fast/forms/color-scheme/select/select-multiple-hover-unselected.html",
			}
			test(failure, `test LIKE "ninja://:blink\\_web\\_tests/%fast/forms/color-scheme/select/select-multiple-hover-unselected.html%"`)
		})
		Convey(`Escaping`, func() {
			failure := &clustering.Failure{
				TestID: `ninja://:blink_web_tests/a/b_\%c.html`,
			}
			test(failure, `test LIKE "ninja://:blink\\_web\\_tests/%a/b\\_\\\\\\%c.html%"`)
		})
	})
	Convey(`Cluster Description`, t, func() {
		a := &Algorithm{}
		cfg, err := compiledcfg.NewConfig(cfgpb)
		So(err, ShouldBeNil)

		Convey(`No matching rules`, func() {
			failure := &clustering.Failure{
				TestID: "ninja://test_name_one",
			}
			description := a.ClusterDescription(cfg, failure)
			So(description.Title, ShouldEqual, "ninja://test_name_one")
			So(description.Description, ShouldContainSubstring, "ninja://test_name_one")
		})
		Convey(`Matching rule`, func() {
			failure := &clustering.Failure{
				TestID: "ninja://:blink_web_tests/virtual/dark-color-scheme/fast/forms/color-scheme/select/select-multiple-hover-unselected.html",
			}
			description := a.ClusterDescription(cfg, failure)
			So(description.Title, ShouldEqual, "ninja://:blink\\_web\\_tests/%fast/forms/color-scheme/select/select-multiple-hover-unselected.html%")
			So(description.Description, ShouldContainSubstring, "ninja://:blink\\_web\\_tests/%fast/forms/color-scheme/select/select-multiple-hover-unselected.html%")
		})
	})
}
