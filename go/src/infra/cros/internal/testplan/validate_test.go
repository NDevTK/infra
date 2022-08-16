package testplan

import (
	"context"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/gerrit"
	"infra/tools/dirmd"
	dirmdpb "infra/tools/dirmd/proto"
	"infra/tools/dirmd/proto/chromeos"

	"go.chromium.org/chromiumos/config/go/test/plan"
)

func TestValidateMapping(t *testing.T) {
	ctx := context.Background()
	testStarlarkContent := "testcontent"
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.NilError(t, ValidateMapping(ctx, client, test.mapping))
		})
	}
}

func TestValidateMappingErrors(t *testing.T) {
	ctx := context.Background()
	client := &gerrit.MockClient{
		T: t,
		ExpectedDownloads: map[gerrit.ExpectedPathParams]*string{
			{
				Host:    "chromium.googlesource.com",
				Project: "testrepo",
				Ref:     "HEAD",
				Path:    "testfile.star",
			}: nil,
		},
	}

	tests := []struct {
		name           string
		mapping        *dirmd.Mapping
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
										PathRegexps: []string{"a/b/.*"},
									},
								},
							},
						},
					},
				},
			},
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
			"all TestPlanStarlarkFile must specify \".star\" files, got \"testfile.txt\"",
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
												Path:    "testfile.star",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"failed downloading file",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateMapping(ctx, client, test.mapping)
			assert.ErrorContains(t, err, test.errorSubstring)
		})
	}
}
