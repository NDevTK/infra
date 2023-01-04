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
	cvpb "go.chromium.org/luci/cv/api/config/v2"
)

// names of relevant ConfigGroups and Builders in the CV config.
const (
	totConfigGroupName          = "ToT"
	cqOrchestratorCvBuilderName = "chromeos/cq/cq-orchestrator"
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

// validateProjectsToCheckInConfigGroup returns an error if any of projectsToCheck
// are not in configGroup.
func validateProjectsToCheckInConfigGroup(configGroup *cvpb.ConfigGroup, projectsToCheck []string) error {
	projectsInConfigGroup := getProjectsFromConfigGroup(configGroup)
	for _, projectToCheck := range projectsToCheck {
		if !projectsInConfigGroup.Has(projectToCheck) {
			return fmt.Errorf("project %q not found in %q ConfigGroup", projectToCheck, configGroup.GetName())
		}
	}

	return nil
}

// validateProjectsToCheckIncludedByBuilder returns an error if any of projectsToCheck
// are not included by builder.
func validateProjectsToCheckIncludedByBuilder(builder *cvpb.Verifiers_Tryjob_Builder, projectsToCheck []string) error {
	for _, projectToCheck := range projectsToCheck {
		included, err := projectIncludedForBuilder(builder, projectToCheck)
		if err != nil {
			return err
		}

		if !included {
			return fmt.Errorf("project %q not included by builder %q", projectToCheck, builder.GetName())
		}
	}

	return nil
}

// getBuilderFromBuildbucketConfig finds builderName in bbCfg, returning an
// error if the builder is not found.
func getBuilderFromBuildbucketConfig(bbCfg *bbpb.BuildbucketCfg, builderName string) (*bbpb.BuilderConfig, error) {
	for _, bucket := range bbCfg.GetBuckets() {
		for _, builder := range bucket.GetSwarming().GetBuilders() {
			if builder.GetName() == builderName {
				return builder, nil
			}
		}
	}

	return nil, fmt.Errorf("no builder named %q found", builderName)
}

// getConfigGroup finds the ConfigGroup with name configGroup in cvConfig,
// returning an error if the ConfigGroup is not found.
func getConfigGroup(cvConfig *cvpb.Config, configGroup string) (*cvpb.ConfigGroup, error) {
	for _, group := range cvConfig.GetConfigGroups() {
		if group.GetName() == configGroup {
			return group, nil
		}
	}

	return nil, fmt.Errorf("ConfigGroup %q not found", configGroup)
}

// getProjectsFromConfigGroup gathers all the project names in configGroup and
// returns them as a set.
func getProjectsFromConfigGroup(configGroup *cvpb.ConfigGroup) stringset.Set {
	result := stringset.New(0)
	for _, gerrit := range configGroup.GetGerrit() {
		for _, project := range gerrit.GetProjects() {
			result.Add(project.GetName())
		}
	}

	return result
}

// getBuilderFromConfigGroup finds the Verifiers_Tryjob_Builder with
// name builderName in cvConfig, returning an error if the
// Verifiers_Tryjob_Builder is not found.
func getBuilderFromConfigGroup(cvConfig *cvpb.ConfigGroup, builderName string) (*cvpb.Verifiers_Tryjob_Builder, error) {
	for _, builder := range cvConfig.GetVerifiers().GetTryjob().GetBuilders() {
		if builder.GetName() == builderName {
			return builder, nil
		}
	}

	return nil, fmt.Errorf("Builder %q not found", builderName)
}

// projectIncludedForBuilder returns whether projectName is included by
// builder's LocationFilters. Only exclude LocationFilters are supported, and
// LocationFilters with project regexp .* are ignored (this function is tailored
// for the current CQ orchestrator CV config).
func projectIncludedForBuilder(builder *cvpb.Verifiers_Tryjob_Builder, projectName string) (bool, error) {
	for _, locationFilter := range builder.GetLocationFilters() {
		if !locationFilter.GetExclude() {
			return false, fmt.Errorf("only exclude LocationFilters currently supported, got %v", locationFilter)
		}

		if locationFilter.GetGerritProjectRegexp() == ".*" {
			continue
		}

		matched, err := regexp.Match(locationFilter.GetGerritProjectRegexp(), []byte(projectName))
		if err != nil {
			return false, err
		}

		if matched {
			return false, nil
		}
	}

	return true, nil
}

// TextSummary returns a string summarizing how many projects in manifest match
// with a ProjectMigrationConfig in bbCfg. If projectsToCheck is non-empty, the
// summary contains whether each of specific project is migrated. Projects that
// are not in the "ToT" ConfigGroup of cvConfig or are excluded from the CQ
// orchestrator by a LocationFilter are skipped.
func TextSummary(
	ctx context.Context,
	manifest *repo.Manifest,
	bbCfg *bbpb.BuildbucketCfg,
	cvConfig *cvpb.Config,
	projectsToCheck []string,
) (string, error) {
	// Create a set version of projectsToCheck, so each project can be checked
	// for membership quickly.
	projectsToCheckSet := stringset.NewFromSlice(projectsToCheck...)
	var summaryBuilder strings.Builder

	totConfigGroup, err := getConfigGroup(cvConfig, totConfigGroupName)
	if err != nil {
		return "", err
	}

	cqOrchCvConfig, err := getBuilderFromConfigGroup(totConfigGroup, cqOrchestratorCvBuilderName)
	if err != nil {
		return "", err
	}

	if err := validateProjectsToCheckInManifest(manifest, projectsToCheck); err != nil {
		return "", err
	}

	if err := validateProjectsToCheckInConfigGroup(totConfigGroup, projectsToCheck); err != nil {
		return "", err
	}

	if err := validateProjectsToCheckIncludedByBuilder(cqOrchCvConfig, projectsToCheck); err != nil {
		return "", err
	}

	totCvProjects := getProjectsFromConfigGroup(totConfigGroup)

	for _, builderName := range orchestratorNames {
		builderConfig, err := getBuilderFromBuildbucketConfig(bbCfg, builderName)
		if err != nil {
			return "", err
		}

		builderProps := &cqOrchProperties{}
		if err := json.Unmarshal([]byte(builderConfig.GetProperties()), builderProps); err != nil {
			return "", fmt.Errorf("error parsing properties for builder %q: %w", builderName, err)
		}

		totalProjects := 0
		projectsMigrated := 0
		projectsToCheckMatchedSet := stringset.New(0)
		for _, project := range manifest.Projects {
			if !totCvProjects.Has(project.Name) {
				logging.Debugf(ctx, "project %q not in ToT ConfigGroup, skipping", project.Name)
				continue
			}

			includedByCqOrch, err := projectIncludedForBuilder(cqOrchCvConfig, project.Name)
			if err != nil {
				return "", err
			}

			if !includedByCqOrch {
				logging.Debugf(ctx, "project %q excluded by CQ orchestrator, skipping", project.Name)
				continue
			}

			totalProjects += 1
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
				builderName, projectsMigrated, totalProjects,
			),
		); err != nil {
			return "", err
		}
	}

	return summaryBuilder.String(), nil
}
