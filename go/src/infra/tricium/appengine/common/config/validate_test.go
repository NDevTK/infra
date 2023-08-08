// Copyright 2017 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"infra/tricium/api/v1"
)

func TestValidate(t *testing.T) {
	Convey("Test Environment", t, func() {
		functionName := "FunctionName"
		platform := tricium.Platform_UBUNTU
		sd := &tricium.ServiceConfig{
			BuildbucketServerHost: "cr-buildbucket-dev.appspot.com",
			Platforms: []*tricium.Platform_Details{
				{
					Name:       platform,
					HasRuntime: true,
				},
			},
			DataDetails: []*tricium.Data_TypeDetails{
				{
					Type:               tricium.Data_GIT_FILE_DETAILS,
					IsPlatformSpecific: false,
				},
				{
					Type:               tricium.Data_RESULTS,
					IsPlatformSpecific: true,
				},
			},
		}
		functions := []*tricium.Function{
			{
				Type:     tricium.Function_ANALYZER,
				Name:     functionName,
				Needs:    tricium.Data_GIT_FILE_DETAILS,
				Provides: tricium.Data_RESULTS,
				Impls: []*tricium.Impl{
					{
						RuntimePlatform:     platform,
						ProvidesForPlatform: platform,
						Impl: &tricium.Impl_Recipe{
							Recipe: &tricium.Recipe{
								Project: "infra",
								Bucket:  "try",
								Builder: "analysis",
							},
						},
					},
				},
			},
		}

		Convey("Supported function platform OK", func() {
			err := Validate(sd, &tricium.ProjectConfig{
				Functions: functions,
				Selections: []*tricium.Selection{
					{
						Function: functionName,
						Platform: platform,
					},
				},
			})
			So(err, ShouldBeNil)
		})

		Convey("Non-supported function platform causes error", func() {
			err := Validate(sd, &tricium.ProjectConfig{
				Functions: functions,
				Selections: []*tricium.Selection{
					{
						Function: functionName,
						Platform: tricium.Platform_WINDOWS,
					},
				},
			})
			So(err, ShouldNotBeNil)
		})
	})
}

func TestMergeFunctions(t *testing.T) {
	Convey("Test Environment", t, func() {
		functionName := "Analyzer"
		platform := tricium.Platform_UBUNTU
		sc := &tricium.ServiceConfig{
			Platforms: []*tricium.Platform_Details{
				{
					Name:       platform,
					HasRuntime: true,
				},
			},
			DataDetails: []*tricium.Data_TypeDetails{
				{
					Type:               tricium.Data_GIT_FILE_DETAILS,
					IsPlatformSpecific: false,
				},
				{
					Type:               tricium.Data_RESULTS,
					IsPlatformSpecific: true,
				},
			},
		}

		Convey("Project function def without service def must have data deps", func() {
			_, err := mergeFunction(functionName, sc, nil, &tricium.Function{
				Type: tricium.Function_ANALYZER,
				Name: functionName,
			})
			So(err, ShouldNotBeNil)
		})

		Convey("Service function def must have data deps", func() {
			_, err := mergeFunction(functionName, sc, &tricium.Function{
				Type: tricium.Function_ANALYZER,
				Name: functionName,
			}, nil)
			So(err, ShouldNotBeNil)
		})

		Convey("No service function config is OK", func() {
			_, err := mergeFunction(functionName, sc, nil, &tricium.Function{
				Type:     tricium.Function_ANALYZER,
				Name:     functionName,
				Needs:    tricium.Data_GIT_FILE_DETAILS,
				Provides: tricium.Data_RESULTS,
			})
			So(err, ShouldBeNil)
		})

		Convey("No project function config is OK", func() {
			_, err := mergeFunction(functionName, sc, &tricium.Function{
				Type:     tricium.Function_ANALYZER,
				Name:     functionName,
				Needs:    tricium.Data_GIT_FILE_DETAILS,
				Provides: tricium.Data_RESULTS,
			}, nil)
			So(err, ShouldBeNil)
		})

		Convey("Change of service data deps not allowed", func() {
			_, err := mergeFunction(functionName, sc, &tricium.Function{
				Type:     tricium.Function_ANALYZER,
				Name:     functionName,
				Needs:    tricium.Data_GIT_FILE_DETAILS,
				Provides: tricium.Data_RESULTS,
			}, &tricium.Function{
				Type:     tricium.Function_ISOLATOR,
				Name:     functionName,
				Provides: tricium.Data_FILES,
			})
			So(err, ShouldNotBeNil)
		})

		Convey("Neither service nor function config not OK", func() {
			_, err := mergeFunction(functionName, sc, nil, nil)
			So(err, ShouldNotBeNil)
		})

		Convey("Project details override service details", func() {
			user := "someone"
			comp := "someonesComp"
			a, err := mergeFunction(functionName, sc, &tricium.Function{
				Type:              tricium.Function_ANALYZER,
				Name:              functionName,
				Needs:             tricium.Data_GIT_FILE_DETAILS,
				Provides:          tricium.Data_RESULTS,
				PathFilters:       []string{"*\\.py", "*\\.pypy"},
				Owner:             "emso",
				MonorailComponent: "compA",
				Impls: []*tricium.Impl{
					{
						ProvidesForPlatform: platform,
						RuntimePlatform:     platform,
						Impl: &tricium.Impl_Recipe{
							Recipe: &tricium.Recipe{
								Project: "infra",
								Bucket:  "try",
								Builder: "analysis",
							},
						},
					},
				},
			}, &tricium.Function{
				Type:              tricium.Function_ANALYZER,
				Name:              functionName,
				PathFilters:       []string{"*\\.py"},
				Owner:             user,
				MonorailComponent: comp,
				Impls: []*tricium.Impl{
					{
						ProvidesForPlatform: platform,
						RuntimePlatform:     platform,
						Impl: &tricium.Impl_Recipe{
							Recipe: &tricium.Recipe{
								Project: "infra",
								Bucket:  "try",
								Builder: "analysis",
							},
						},
					},
				},
			})
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			So(a.Owner, ShouldEqual, user)
			So(a.MonorailComponent, ShouldEqual, comp)
			So(len(a.PathFilters), ShouldEqual, 1)
			So(len(a.Impls), ShouldEqual, 1)
		})
	})
}

func TestMergeImpls(t *testing.T) {
	Convey("Test Environment", t, func() {
		si := []*tricium.Impl{
			{
				ProvidesForPlatform: tricium.Platform_UBUNTU,
			},
			{
				ProvidesForPlatform: tricium.Platform_WINDOWS,
			},
		}
		pi := []*tricium.Impl{
			{
				ProvidesForPlatform: tricium.Platform_WINDOWS,
			},
			{
				ProvidesForPlatform: tricium.Platform_MAC,
			},
		}
		Convey("Merges impls with override", func() {
			mi := mergeImpls(si, pi)
			So(len(mi), ShouldEqual, 3)
		})
	})
}
