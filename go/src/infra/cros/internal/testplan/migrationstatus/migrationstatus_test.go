package migrationstatus_test

import (
	"context"
	"testing"

	"infra/cros/internal/assert"
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
								},
								{
									"project": "chromeos/a/b"
								},
								{
									"project": "chromeos/c/d"
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
							Name: "chromeos/c/d",
						},
						{
							Name: "chromemos/e/f",
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
			Name: "chromeos/c/d",
		},
		{
			Name: "chromemos/e/f",
		},
		{
			Name: "chromeos/projectnotincvconfig",
		},
		{
			Name: "chromeos/excludedbyorch1",
		},
	},
}

func TestTextSummary(t *testing.T) {
	ctx := context.Background()

	t.Run("no projectsToCheck", func(t *testing.T) {
		summary, err := migrationstatus.TextSummary(ctx, manifest, bbCfg, cvConfig, []string{})
		if err != nil {
			t.Fatalf("TextSummary returned error: %s", err)
		}

		expectedSummary := `cq-orchestrator: 2 / 4 projects migrated
staging-cq-orchestrator: 3 / 4 projects migrated
`
		if diff := cmp.Diff(expectedSummary, summary); diff != "" {
			t.Errorf("TextSummary returned unexpected summary (-want +got):\n%s", diff)
		}
	})

	t.Run("projectsToCheck", func(t *testing.T) {
		summary, err := migrationstatus.TextSummary(ctx, manifest, bbCfg, cvConfig, []string{"chromeos/c/d"})
		if err != nil {
			t.Fatalf("TextSummary returned error: %s", err)
		}

		expectedSummary := `cq-orchestrator: project chromeos/c/d not migrated
cq-orchestrator: 2 / 4 projects migrated
staging-cq-orchestrator: project chromeos/c/d migrated
staging-cq-orchestrator: 3 / 4 projects migrated
`
		if diff := cmp.Diff(expectedSummary, summary); diff != "" {
			t.Errorf("TextSummary returned unexpected summary (-want +got):\n%s", diff)
		}
	})
}

func TestTextSummaryErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("projectToCheck missing from manifest", func(t *testing.T) {
		_, err := migrationstatus.TextSummary(ctx, manifest, bbCfg, cvConfig, []string{"missingproject"})
		assert.ErrorContains(t, err, `project "missingproject" not found in manifest`)
	})

	t.Run("projectToCheck missing from CqGroup", func(t *testing.T) {
		_, err := migrationstatus.TextSummary(ctx, manifest, bbCfg, cvConfig, []string{"chromeos/projectnotincvconfig"})
		assert.ErrorContains(t, err, `project "chromeos/projectnotincvconfig" not found in "ToT" ConfigGroup`)
	})

	t.Run("projectToCheck excluded by cq orch", func(t *testing.T) {
		_, err := migrationstatus.TextSummary(ctx, manifest, bbCfg, cvConfig, []string{"chromeos/excludedbyorch1"})
		assert.ErrorContains(t, err, `project "chromeos/excludedbyorch1" not included by builder "chromeos/cq/cq-orchestrator"`)
	})
}
