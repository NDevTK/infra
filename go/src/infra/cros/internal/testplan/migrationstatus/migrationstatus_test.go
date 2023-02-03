package migrationstatus_test

import (
	"context"
	"testing"

	"infra/cros/internal/repo"
	"infra/cros/internal/testplan/migrationstatus"

	"github.com/google/go-cmp/cmp"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	cvpb "go.chromium.org/luci/cv/api/config/v2"
)

var bbCfg *bbpb.BuildbucketCfg = &bbpb.BuildbucketCfg{
	Buckets: []*bbpb.Bucket{
		{
			Swarming: &bbpb.Swarming{
				Builders: []*bbpb.BuilderConfig{
					{
						Name: "amd64-generic-cq",
					},
					{
						Name: "cq-orchestrator",
						Properties: `{
						"$chromeos/cros_test_plan_v2": {
							"migration_configs": [
								{
									"project": "chromeos/testprojects/.*",
									"other_field": 123
								},
								{
									"project": "chromeos/a/b"
								}
							]
						}
						}`,
					},
					{
						Name: "staging-cq-orchestrator",
						Properties: `{
						"$chromeos/cros_test_plan_v2": {
							"migration_configs": [
								{
									"project": "chromeos/testprojects/.*"
								}
							]
						}
						}`,
					},
				},
			},
		},
	},
}

var cvConfig = &cvpb.Config{
	ConfigGroups: []*cvpb.ConfigGroup{
		{
			Name: "ToT",
			Gerrit: []*cvpb.ConfigGroup_Gerrit{
				{
					Projects: []*cvpb.ConfigGroup_Gerrit_Project{
						{
							Name: "chromeos/testprojects/testproject1",
						},
						{
							Name: "chromeos/a/b",
						},
						{
							Name: "chromeos/excludedbyorch1",
						},
					},
				},
			},
			Verifiers: &cvpb.Verifiers{
				Tryjob: &cvpb.Verifiers_Tryjob{
					Builders: []*cvpb.Verifiers_Tryjob_Builder{
						{
							Name: "chromeos/cq/cq-orchestrator",
							LocationFilters: []*cvpb.Verifiers_Tryjob_Builder_LocationFilter{
								{
									Exclude:             true,
									GerritProjectRegexp: "chromeos/excludedbyorch.*",
								},
							},
						},
					},
				},
			},
		},
	},
}

var manifest *repo.Manifest = &repo.Manifest{
	Projects: []repo.Project{
		{
			Name: "chromeos/testprojects/testproject1",
		},
		{
			Name: "chromeos/a/b",
		},
		{
			Name: "chromeos/projectnotincvconfig",
		},
		{
			Name: "chromeos/excludedbyorch1",
		},
	},
}

func TestCompute(t *testing.T) {
	ctx := context.Background()

	statuses, err := migrationstatus.Compute(ctx, manifest, bbCfg, cvConfig)
	if err != nil {
		t.Fatalf("TextSummary returned error: %s", err)
	}

	expected := []*migrationstatus.MigrationStatus{
		// testproject1 is included by regex match.
		{
			BuilderName:            "cq-orchestrator",
			ProjectName:            "chromeos/testprojects/testproject1",
			MatchesMigrationConfig: true,
			IncludedByToT:          true,
			IncludedByBuilder:      true,
		},
		{
			BuilderName:            "cq-orchestrator",
			ProjectName:            "chromeos/a/b",
			MatchesMigrationConfig: true,
			IncludedByToT:          true,
			IncludedByBuilder:      true,
		},
		{
			BuilderName:            "cq-orchestrator",
			ProjectName:            "chromeos/projectnotincvconfig",
			MatchesMigrationConfig: false,
			IncludedByToT:          false,
			IncludedByBuilder:      true,
		},
		{
			BuilderName:            "cq-orchestrator",
			ProjectName:            "chromeos/excludedbyorch1",
			MatchesMigrationConfig: false,
			IncludedByToT:          true,
			IncludedByBuilder:      false,
		},
		{
			BuilderName:            "staging-cq-orchestrator",
			ProjectName:            "chromeos/testprojects/testproject1",
			MatchesMigrationConfig: true,
			IncludedByToT:          true,
			IncludedByBuilder:      true,
		},
		{
			BuilderName:            "staging-cq-orchestrator",
			ProjectName:            "chromeos/a/b",
			MatchesMigrationConfig: false,
			IncludedByToT:          true,
			IncludedByBuilder:      true,
		},
		{
			BuilderName:            "staging-cq-orchestrator",
			ProjectName:            "chromeos/projectnotincvconfig",
			MatchesMigrationConfig: false,
			IncludedByToT:          false,
			IncludedByBuilder:      true,
		},
		{
			BuilderName:            "staging-cq-orchestrator",
			ProjectName:            "chromeos/excludedbyorch1",
			MatchesMigrationConfig: false,
			IncludedByToT:          true,
			IncludedByBuilder:      false,
		},
	}

	if diff := cmp.Diff(expected, statuses); diff != "" {
		t.Errorf("Compute returned unexpected statuses (-want +got):\n%s", diff)
	}
}

func TestTextSummary(t *testing.T) {
	ctx := context.Background()

	statuses := []*migrationstatus.MigrationStatus{
		{
			BuilderName:            "cq-orchestrator",
			ProjectName:            "projectA",
			MatchesMigrationConfig: false,
			IncludedByToT:          true,
			IncludedByBuilder:      true,
		},
		{
			BuilderName:            "staging-cq-orchestrator",
			ProjectName:            "projectA",
			MatchesMigrationConfig: true,
			IncludedByToT:          true,
			IncludedByBuilder:      true,
		},
		{
			BuilderName:            "cq-orchestrator",
			ProjectName:            "projectB",
			MatchesMigrationConfig: true,
			IncludedByToT:          true,
			IncludedByBuilder:      true,
		},
		{
			BuilderName:            "staging-cq-orchestrator",
			ProjectName:            "projectB",
			MatchesMigrationConfig: true,
			IncludedByToT:          true,
			IncludedByBuilder:      true,
		},
		{
			BuilderName:            "cq-orchestrator",
			ProjectName:            "projectC",
			MatchesMigrationConfig: false,
			IncludedByToT:          false,
			IncludedByBuilder:      false,
		},
		{
			BuilderName:            "staging-cq-orchestrator",
			ProjectName:            "projectC",
			MatchesMigrationConfig: false,
			IncludedByToT:          false,
			IncludedByBuilder:      false,
		},
		{
			BuilderName:            "cq-orchestrator",
			ProjectName:            "projectD",
			MatchesMigrationConfig: false,
			IncludedByToT:          true,
			IncludedByBuilder:      false,
		},
		{
			BuilderName:            "staging-cq-orchestrator",
			ProjectName:            "projectD",
			MatchesMigrationConfig: false,
			IncludedByToT:          true,
			IncludedByBuilder:      false,
		},
	}

	summary, err := migrationstatus.TextSummary(ctx, statuses, []string{"projectA", "projectC", "projectD"})
	if err != nil {
		t.Fatalf("TextSummary returned error: %s", err)
	}

	expectedSummary := `cq-orchestrator: project projectA not migrated
cq-orchestrator: project projectC not included in "ToT" ConfigGroup
cq-orchestrator: project projectD not included by builder
cq-orchestrator: 1 / 2 projects migrated
staging-cq-orchestrator: project projectA migrated
staging-cq-orchestrator: project projectC not included in "ToT" ConfigGroup
staging-cq-orchestrator: project projectD not included by builder
staging-cq-orchestrator: 2 / 2 projects migrated
`
	if diff := cmp.Diff(expectedSummary, summary); diff != "" {
		t.Errorf("TextSummary returned unexpected summary (-want +got):\n%s", diff)
	}
}
