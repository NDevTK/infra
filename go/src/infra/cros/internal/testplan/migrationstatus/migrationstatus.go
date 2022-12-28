// package migrationstatus summarizes the status of projects being migrated to
// distributed test config.
package migrationstatus

import (
	"context"
	"encoding/json"
	"fmt"
	"infra/cros/internal/repo"
	"regexp"
	"strings"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/logging"
)

// orchestratorNames is a list of builder names with ProjectMigrationStatus
// input properties.
var orchestratorNames = []string{"cq-orchestrator", "staging-cq-orchestrator"}

// cqOrchProperties is used to parse the JSON input properties of orchestrators,
// must be kept in sync with https://source.chromium.org/chromium/chromiumos/infra/recipes/+/main:recipe_modules/cros_test_plan_v2/cros_test_plan_v2.proto.
type cqOrchProperties struct {
	CrosTestPlanV2 struct {
		MigrationConfigs []struct {
			Project string
		} `json:"migration_configs"`
	} `json:"$chromeos/cros_test_plan_v2"`
}

// validateProjectsToCheckInManifest returns an error if any of projectsToCheck
// are not in manifest.
func validateProjectsToCheckInManifest(manifest *repo.Manifest, projectsToCheck []string) error {
	manifestProjectsSet := stringset.New(0)
	for _, project := range manifest.Projects {
		manifestProjectsSet.Add(project.Name)
	}

	for _, projectToCheck := range projectsToCheck {
		if !manifestProjectsSet.Has(projectToCheck) {
			return fmt.Errorf("project %q not found in manifest", projectToCheck)
		}
	}

	return nil
}

// TextSummary returns a string summarizing how many projects in manifest match
// with a ProjectMigrationConfig in bbCfg. If projectsToCheck is non-empty, the
// summary contains whether each of specific project is migrated.
func TextSummary(
	ctx context.Context,
	manifest *repo.Manifest,
	bbCfg *bbpb.BuildbucketCfg,
	projectsToCheck []string,
) (string, error) {
	// Create a set version of projectsToCheck, so each project can be checked
	// for membership quickly.
	projectsToCheckSet := stringset.NewFromSlice(projectsToCheck...)
	var summaryBuilder strings.Builder

	if err := validateProjectsToCheckInManifest(manifest, projectsToCheck); err != nil {
		return "", err
	}

	for _, builderName := range orchestratorNames {
		// Find the BuilderConfig for builderName, return an error if not found.
		var builderConfig *bbpb.BuilderConfig
		for _, bucket := range bbCfg.Buckets {
			for _, builder := range bucket.GetSwarming().GetBuilders() {
				if builder.GetName() == builderName {
					builderConfig = builder
					break
				}
			}
		}

		if builderConfig == nil {
			return "", fmt.Errorf("no builder named %q found", builderName)
		}

		builderProps := &cqOrchProperties{}
		if err := json.Unmarshal([]byte(builderConfig.GetProperties()), builderProps); err != nil {
			return "", fmt.Errorf("error parsing properties for builder %q: %w", builderName, err)
		}

		projectsMigrated := 0
		projectsToCheckMatchedSet := stringset.New(0)
		for _, project := range manifest.Projects {
			for _, migrationConfig := range builderProps.CrosTestPlanV2.MigrationConfigs {
				// Empty project name in the MigrationConfig will regexp match
				// all strings, and is unexpected.
				if migrationConfig.Project == "" {
					return "", fmt.Errorf("unexpected MigrationConfig with empty project: %q", migrationConfig)
				}

				matched, err := regexp.Match(migrationConfig.Project, []byte(project.Name))
				if err != nil {
					return "", err
				}

				// If project matches any MigrationConfig, add it to the count
				// of projects migrated. Also see if it is in the set of
				// projects to specifically check the status of.
				if matched {
					logging.Debugf(ctx, "matched project %q with migration config %q", project, migrationConfig)
					projectsMigrated += 1

					if projectsToCheckSet.Has(project.Name) {
						projectsToCheckMatchedSet.Add(project.Name)
					}

					break
				}
			}
		}

		// Add the status of projectsToCheck to the summary.
		for _, project := range projectsToCheckSet.ToSortedSlice() {
			if projectsToCheckMatchedSet.Has(project) {
				summaryBuilder.WriteString(fmt.Sprintf("%s: project %s migrated\n", builderName, project))
			} else {
				summaryBuilder.WriteString(fmt.Sprintf("%s: project %s not migrated\n", builderName, project))
			}
		}

		// Write the summary of projects migrated for the builder.
		if _, err := summaryBuilder.WriteString(
			fmt.Sprintf(
				"%s: %d / %d projects migrated\n",
				builderName, projectsMigrated, len(manifest.Projects),
			),
		); err != nil {
			return "", err
		}
	}

	return summaryBuilder.String(), nil
}
