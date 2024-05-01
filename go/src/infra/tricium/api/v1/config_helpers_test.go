// Copyright 2017 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tricium

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/auth/identity"
	"go.chromium.org/luci/common/logging/memlogger"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"
)

func TestLookupRepoDetails(t *testing.T) {

	pc := &ProjectConfig{
		Repos: []*RepoDetails{
			{
				Source: &RepoDetails_GitRepo{
					GitRepo: &GitRepo{
						Url: "https://github.com/google/gitiles.git",
					},
				},
			},
			{
				Source: &RepoDetails_GerritProject{
					GerritProject: &GerritProject{
						Host:    "chromium.googlesource.com",
						Project: "infra/infra",
						GitUrl:  "https://chromium.googlesource.com/infra/infra.git",
					},
				},
			},
		},
	}

	Convey("Matches GerritProject when URL matches", t, func() {
		request := &AnalyzeRequest{
			Source: &AnalyzeRequest_GerritRevision{
				GerritRevision: &GerritRevision{
					GitUrl: "https://chromium.googlesource.com/infra/infra.git",
					GitRef: "refs/changes/97/12397/1",
				},
			},
		}
		So(LookupRepoDetails(pc, request), ShouldEqual, pc.Repos[1])
	})

	Convey("Matches GitRepo when URL matches", t, func() {
		request := &AnalyzeRequest{
			Source: &AnalyzeRequest_GitCommit{
				GitCommit: &GitCommit{
					Url: "https://github.com/google/gitiles.git",
					Ref: "refs/heads/master",
				},
			},
		}
		So(LookupRepoDetails(pc, request), ShouldEqual, pc.Repos[0])
	})

	Convey("Returns nil when no repo is found", t, func() {
		request := &AnalyzeRequest{
			Source: &AnalyzeRequest_GerritRevision{
				GerritRevision: &GerritRevision{
					GitUrl: "https://foo.googlesource.com/bar",
					GitRef: "refs/changes/97/197/2",
				},
			},
		}
		So(LookupRepoDetails(pc, request), ShouldBeNil)
	})
}

func TestCanRequest(t *testing.T) {
	ctx := memory.Use(memlogger.Use(context.Background()))

	okACLGroup := "tricium-playground-requesters"
	okACLUser := "user:ok@example.com"
	pc := &ProjectConfig{
		Acls: []*Acl{
			{
				Role:  Acl_REQUESTER,
				Group: okACLGroup,
			},
			{
				Role:     Acl_REQUESTER,
				Identity: okACLUser,
			},
		},
	}

	Convey("Only users in OK ACL group can request", t, func() {
		ctx = auth.WithState(ctx, &authtest.FakeState{
			Identity:       "user:abc@example.com",
			IdentityGroups: []string{okACLGroup},
		})
		ok, err := CanRequest(ctx, pc)
		So(err, ShouldBeNil)
		So(ok, ShouldBeTrue)
	})

	Convey("User with OK ACL can request", t, func() {
		ctx = auth.WithState(ctx, &authtest.FakeState{
			Identity: identity.Identity(okACLUser),
		})
		ok, err := CanRequest(ctx, pc)
		So(err, ShouldBeNil)
		So(ok, ShouldBeTrue)
	})

	Convey("Anonymous users cannot request", t, func() {
		ctx = auth.WithState(ctx, &authtest.FakeState{
			Identity: identity.AnonymousIdentity,
		})
		ok, err := CanRequest(ctx, pc)
		So(err, ShouldBeNil)
		So(ok, ShouldBeFalse)
	})
}

func TestLookupFunction(t *testing.T) {

	functions := []*Function{
		{
			Name: "Pylint",
			Type: Function_ANALYZER,
		},
	}

	Convey("Known function is known", t, func() {
		So(LookupFunction(functions, "Pylint"), ShouldResemble, functions[0])
	})

	Convey("Unknown function is unknown", t, func() {
		So(LookupFunction(functions, "blabla"), ShouldBeNil)
	})
}

func TestSupportsPlatform(t *testing.T) {
	analyzer := &Function{
		Type: Function_ANALYZER,
		Name: "PyLint",
		Impls: []*Impl{
			{
				ProvidesForPlatform: Platform_WINDOWS,
			},
			{
				ProvidesForPlatform: Platform_UBUNTU,
			},
		},
	}

	Convey("Supported platform is supported", t, func() {
		So(SupportsPlatform(analyzer, Platform_UBUNTU), ShouldBeTrue)
	})

	Convey("Unsupported platform is not supported", t, func() {
		So(SupportsPlatform(analyzer, Platform_MAC), ShouldBeFalse)
	})

	Convey("ANY platform always supported", t, func() {
		So(SupportsPlatform(analyzer, Platform_ANY), ShouldBeTrue)
	})
}

func TestLookupImplForPlatform(t *testing.T) {
	implForLinux := &Impl{ProvidesForPlatform: Platform_LINUX}
	implForMac := &Impl{ProvidesForPlatform: Platform_MAC}
	analyzer := &Function{
		Impls: []*Impl{
			implForLinux,
			implForMac,
		},
	}

	Convey("Impl for known platform is returned", t, func() {
		i := LookupImplForPlatform(analyzer, Platform_LINUX)
		So(i, ShouldEqual, implForLinux)
	})

	Convey("Impl for any platform returns first", t, func() {
		// In this case, there is no implementation in
		// the list that is explicitly for any platform;
		// we return the first implementation.
		i := LookupImplForPlatform(analyzer, Platform_ANY)
		So(i, ShouldEqual, implForLinux)
	})

	Convey("Impl for unknown platform returns nil", t, func() {
		i := LookupImplForPlatform(analyzer, Platform_WINDOWS)
		So(i, ShouldBeNil)
	})

	implForAny := &Impl{ProvidesForPlatform: Platform_ANY}
	analyzer = &Function{
		Impls: []*Impl{
			implForLinux,
			implForAny,
		},
	}

	Convey("Impl for 'any' platform is used if present", t, func() {
		// In this case, there is an implementation in
		// the list that is explicitly for any platform;
		// we return the 'any' implementation.
		i := LookupImplForPlatform(analyzer, Platform_ANY)
		So(i, ShouldEqual, implForAny)
	})
}

func TestLookupPlatform(t *testing.T) {
	platform := Platform_UBUNTU
	sc := &ServiceConfig{Platforms: []*Platform_Details{{Name: platform}}}

	Convey("Known platform is returned", t, func() {
		p := LookupPlatform(sc, platform)
		So(p, ShouldNotBeNil)
	})

	Convey("Unknown platform returns nil", t, func() {
		p := LookupPlatform(sc, Platform_WINDOWS)
		So(p, ShouldBeNil)
	})
}

func TestValidateFunction(t *testing.T) {

	sc := &ServiceConfig{
		Platforms: []*Platform_Details{
			{
				Name:       Platform_LINUX,
				Dimensions: []string{"pool:Default"},
				HasRuntime: true,
			},
			{
				Name:       Platform_IOS,
				Dimensions: []string{"pool:Default"},
				HasRuntime: false,
			},
		},
	}

	Convey("Function with all required fields is valid", t, func() {
		f := &Function{
			Type:     Function_ANALYZER,
			Name:     "PyLint",
			Needs:    Data_FILES,
			Provides: Data_RESULTS,
		}
		So(ValidateFunction(f, sc), ShouldBeNil)
	})

	Convey("Function names must not be non-empty", t, func() {
		f := &Function{
			Type:     Function_ANALYZER,
			Name:     "",
			Needs:    Data_FILES,
			Provides: Data_RESULTS,
		}
		So(ValidateFunction(f, sc), ShouldNotBeNil)
	})

	Convey("Function names must not contain underscore", t, func() {
		f := &Function{
			Type:     Function_ANALYZER,
			Name:     "Py_Lint",
			Needs:    Data_FILES,
			Provides: Data_RESULTS,
		}
		So(ValidateFunction(f, sc), ShouldNotBeNil)
	})

	Convey("Function without type is invalid", t, func() {
		f := &Function{
			Name:     "PyLint",
			Needs:    Data_FILES,
			Provides: Data_RESULTS,
		}
		So(ValidateFunction(f, sc), ShouldNotBeNil)
	})

	Convey("Function without name is invalid", t, func() {
		f := &Function{
			Type:     Function_ANALYZER,
			Needs:    Data_FILES,
			Provides: Data_RESULTS,
		}
		So(ValidateFunction(f, sc), ShouldNotBeNil)
	})

	Convey("Analyzer function must return results", t, func() {
		f := &Function{
			Type:     Function_ANALYZER,
			Name:     "ConfusedAnalyzer",
			Needs:    Data_FILES,
			Provides: Data_GIT_FILE_DETAILS,
		}
		So(ValidateFunction(f, sc), ShouldNotBeNil)
		f.Provides = Data_RESULTS
		So(ValidateFunction(f, sc), ShouldBeNil)
	})

	Convey("Function with impl without platforms is invalid", t, func() {
		f := &Function{
			Type:  Function_ANALYZER,
			Name:  "PyLint",
			Impls: []*Impl{{}},
		}
		So(ValidateFunction(f, sc), ShouldNotBeNil)
	})
}

func TestValidateImpl(t *testing.T) {

	sc := &ServiceConfig{
		Platforms: []*Platform_Details{
			{
				Name:       Platform_UBUNTU,
				Dimensions: []string{"pool:Default"},
				HasRuntime: true,
			},
			{
				Name:       Platform_ANDROID,
				HasRuntime: false,
			},
		},
	}

	anyType := &Data_TypeDetails{
		IsPlatformSpecific: false,
	}

	Convey("Impl must have a recipe specified", t, func() {
		impl := &Impl{
			RuntimePlatform: Platform_UBUNTU,
		}
		So(validateImpl(impl, sc, anyType, anyType), ShouldNotBeNil)
	})
}
