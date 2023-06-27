// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Some expected error strings are filesystem-specific, skip on windows.
//go:build !windows

package testplan

import (
	"context"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/gerrit"
	"infra/tools/dirmd"
	dirmdpb "infra/tools/dirmd/proto"
	"infra/tools/dirmd/proto/chromeos"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/config/go/test/plan"
)

func TestValidateMapping(t *testing.T) {
	ctx := context.Background()
	testStarlarkContent := "testcontent"
	templatedStarlarkContent := `testplan.get_suite_name()`
	client := &gerrit.MockClient{
		T: t,
		ExpectedDownloads: map[gerrit.ExpectedPathParams]*string{
			{
				Host:    "chromium.googlesource.com",
				Project: "test/repo",
				Ref:     "HEAD",
				Path:    "a/b/c/test.star",
			}: &testStarlarkContent,
			{
				Host:    "chromium.googlesource.com",
				Project: "test/repo1",
				Ref:     "HEAD",
				Path:    "a/b/c/test.star",
			}: &testStarlarkContent,
			{
				Host:    "chromium.googlesource.com",
				Project: "test/repo2",
				Ref:     "HEAD",
				Path:    "test2.star",
			}: &testStarlarkContent,
			{
				Host:    "chromium.googlesource.com",
				Project: "test/repo",
				Ref:     "HEAD",
				Path:    "templated.star",
			}: &templatedStarlarkContent,
		},
	}

	tests := []struct {
		name    string
		mapping *dirmd.Mapping
	}{
		{
			"no ChromeOS metadata",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b/c": {
						TeamEmail: "exampleteam@google.com",
					},
				},
			},
		},
		{
			"single starlark file",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b/c": {
						TeamEmail: "exampleteam@google.com",
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "test/repo",
												Path:    "a/b/c/test.star",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"multiple starlark files",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b/c": {
						TeamEmail: "exampleteam@google.com",
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "test/repo1",
												Path:    "a/b/c/test.star",
											},
											{
												Host:    "chromium.googlesource.com",
												Project: "test/repo2",
												Path:    "test2.star",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"valid regexps",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b/c": {
						TeamEmail: "exampleteam@google.com",
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "test/repo",
												Path:    "a/b/c/test.star",
											},
										},
										PathRegexps:        []string{"a/b/c/d/.*"},
										PathRegexpExcludes: []string{`a/b/c/.*\.md`},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"root directory",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					".": {
						TeamEmail: "exampleteam@google.com",
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "test/repo",
												Path:    "a/b/c/test.star",
											},
										},
										PathRegexps:        []string{"a/b/c/d/.*"},
										PathRegexpExcludes: []string{`a/b/c/.*\.md`},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"regexp doesn't match file",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b": {
						TeamEmail: "exampleteam@google.com",
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "test/repo",
												Path:    "a/b/c/test.star",
											},
										},
										// This doesn't match anything under
										// ./testdata/good_dirmd. This isn't an
										// error, but a warning is logged.
										PathRegexps: []string{"a/b/c/d/nomatch.*"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"template parameters",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b": {
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "test/repo",
												Path:    "templated.star",
												TemplateParameters: &plan.SourceTestPlan_TestPlanStarlarkFile_TemplateParameters{
													SuiteName: "mysuiteA",
													TagCriteria: &api.TestSuite_TestCaseTagCriteria{
														Tags:        []string{"group:mygroupA"},
														TagExcludes: []string{"informational"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.NilError(t, ValidateMapping(ctx, client, test.mapping, "./testdata/good_dirmd"))
		})
	}
}

func TestValidateMappingErrors(t *testing.T) {
	ctx := context.Background()
	testfileContents := "print('hello')"
	client := &gerrit.MockClient{
		T: t,
		ExpectedDownloads: map[gerrit.ExpectedPathParams]*string{
			{
				Host:    "chromium.googlesource.com",
				Project: "testrepo",
				Ref:     "HEAD",
				Path:    "missingtestfile.star",
			}: nil,
			{
				Host:    "chromium.googlesource.com",
				Project: "testrepo",
				Ref:     "HEAD",
				Path:    "testfile.txt",
			}: nil,
			{
				Host:    "chromium.googlesource.com",
				Project: "testrepo",
				Ref:     "HEAD",
				Path:    "testfile.star",
			}: &testfileContents,
			{
				Host:    "chromium.googlesource.com",
				Project: "test/repo",
				Ref:     "HEAD",
				Path:    "templated.star",
			}: &testfileContents,
		},
	}

	tests := []struct {
		name           string
		mapping        *dirmd.Mapping
		repoRoot       string
		errorSubstring string
	}{
		{
			"starlark files empty",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b/c": {
						TeamEmail: "exampleteam@google.com",
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										PathRegexps: []string{"a/b/c/.*"},
									},
								},
							},
						},
					},
				},
			},
			"./testdata/good_dirmd",
			"at least one TestPlanStarlarkFile must be specified",
		},
		{
			"invalid regexp",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b/c": {
						TeamEmail: "exampleteam@google.com",
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "testrepo",
												Path:    "testfile.star",
											},
										},
										PathRegexps: []string{"a/b/c/d/["},
									},
								},
							},
						},
					},
				},
			},
			"./testdata/good_dirmd",
			"failed to compile path regexp",
		},
		{
			"invalid regexp prefix",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b/c": {
						TeamEmail: "exampleteam@google.com",
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "testrepo",
												Path:    "testfile.star",
											},
										},
										PathRegexps: []string{`a/b/e/.*\.txt`},
									},
								},
							},
						},
					},
				},
			},
			"./testdata/good_dirmd",
			"path_regexp(_exclude)s defined in a directory that is not the root of the repo must have the sub-directory as a prefix",
		},
		{
			"invalid file type",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					".": {Chromeos: &chromeos.ChromeOS{
						Cq: &chromeos.ChromeOS_CQ{
							SourceTestPlans: []*plan.SourceTestPlan{
								{
									TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
										{
											Host:    "chromium.googlesource.com",
											Project: "testrepo",
											Path:    "testfile.txt",
										},
									},
								},
							},
						},
					},
					},
				},
			},
			"./testdata/good_dirmd",
			"all TestPlanStarlarkFile must specify \".star\" files, got \"testfile.txt\" (and 1 other error)",
		},
		{
			"starlark file missing",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					".": {
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "testrepo",
												Path:    "missingtestfile.star",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"./testdata/good_dirmd",
			"failed downloading file",
		},
		{
			"non-existant repo root",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					".": {
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "testrepo",
												Path:    "testfile.star",
											},
										},
										PathRegexps: []string{`a/b/e/.*\.txt`},
									},
								},
							},
						},
					},
				},
			},
			"badreporoot",
			"lstat badreporoot: no such file or directory",
		},
		{
			"suite name missing",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b": {
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "test/repo",
												Path:    "templated.star",
												TemplateParameters: &plan.SourceTestPlan_TestPlanStarlarkFile_TemplateParameters{
													TagCriteria: &api.TestSuite_TestCaseTagCriteria{
														Tags:        []string{"group:mygroupA"},
														TagExcludes: []string{"informational"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"./testdata/good_dirmd",
			"suite_name must not be empty",
		},
		{
			"informational not excluded",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b": {
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "test/repo",
												Path:    "templated.star",
												TemplateParameters: &plan.SourceTestPlan_TestPlanStarlarkFile_TemplateParameters{
													SuiteName: "mysuiteA",
													TagCriteria: &api.TestSuite_TestCaseTagCriteria{
														Tags: []string{"group:mygroupA"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"./testdata/good_dirmd",
			`tag_excludes must exclude "informational"`,
		},
		{
			"not templated file",
			&dirmd.Mapping{
				Dirs: map[string]*dirmdpb.Metadata{
					"a/b": {
						Chromeos: &chromeos.ChromeOS{
							Cq: &chromeos.ChromeOS_CQ{
								SourceTestPlans: []*plan.SourceTestPlan{
									{
										TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
											{
												Host:    "chromium.googlesource.com",
												Project: "test/repo",
												Path:    "templated.star",
												TemplateParameters: &plan.SourceTestPlan_TestPlanStarlarkFile_TemplateParameters{
													SuiteName: "mysuiteA",
													TagCriteria: &api.TestSuite_TestCaseTagCriteria{
														Tags:        []string{"group:mygroupA"},
														TagExcludes: []string{"informational"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"./testdata/good_dirmd",
			"setting TemplateParameters has no effect",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateMapping(ctx, client, test.mapping, test.repoRoot)
			assert.ErrorContains(t, err, test.errorSubstring)
		})
	}
}
