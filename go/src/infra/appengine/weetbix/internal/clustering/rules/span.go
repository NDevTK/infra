// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rules

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"time"

	"cloud.google.com/go/spanner"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/bugs"
	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/clustering/rules/lang"
	"infra/appengine/weetbix/internal/config"
	spanutil "infra/appengine/weetbix/internal/span"
)

// RuleIDRe is the regular expression pattern that matches validly
// formed rule IDs.
const RuleIDRePattern = `[0-9a-f]{32}`

// RuleIDRe matches validly formed rule IDs.
var RuleIDRe = regexp.MustCompile(`^` + RuleIDRePattern + `$`)

// UserRe matches valid users. These are email addresses or the special
// value "weetbix".
var UserRe = regexp.MustCompile(`^weetbix|([a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+)$`)

// WeetbixSystem is the special user that identifies changes made by the
// Weetbix system itself in audit fields.
const WeetbixSystem = "weetbix"

// StartingEpoch is the rule last updated time used for projects that have
// no rules (active or otherwise). It is deliberately different from the
// timestamp zero value to be discernible from "timestamp not populated"
// programming errors.
var StartingEpoch = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

// StartingEpoch is the rule version used for projects that have
// no rules (active or otherwise).
var StartingVersion = Version{
	Predicates: StartingEpoch,
	Total:      StartingEpoch,
}

// NotExistsErr is returned by Read methods for a single failure
// association rule, if no matching rule exists.
var NotExistsErr = errors.New("no matching rule exists")

// FailureAssociationRule associates failures with a bug. When the rule
// is used to match incoming test failures, the resultant cluster is
// known as a 'bug cluster' because the cluster is associated with a bug
// (via the failure association rule).
type FailureAssociationRule struct {
	// The LUCI Project for which this rule is defined.
	Project string `json:"project"`
	// The unique identifier for the failure association rule,
	// as 32 lowercase hexadecimal characters.
	RuleID string `json:"ruleId"`
	// The rule predicate, defining which failures are being associated.
	RuleDefinition string `json:"ruleDefinition"`
	// The time the rule was created. Output only.
	CreationTime time.Time `json:"creationTime"`
	// The user which created the rule. Output only.
	CreationUser string `json:"creationUser"`
	// The time the rule was last updated. Output only.
	LastUpdated time.Time `json:"lastUpdated"`
	// The user which last updated the rule. Output only.
	LastUpdatedUser string `json:"lastUpdatedUser"`
	// The time the rule was last updated in a way that caused the
	// matched failures to change, i.e. because of a change to RuleDefinition
	// or IsActive. (By contrast, updating BugID does NOT change
	// the matched failures, so does NOT update this field.)
	// When this value changes, it triggers re-clustering.
	// Compare with RulesVersion on ReclusteringRuns to identify
	// reclustering state.
	// Output only.
	PredicateLastUpdated time.Time `json:"predicateLastUpdated"`
	// BugID is the identifier of the bug that the failures are
	// associated with.
	BugID bugs.BugID `json:"bugId"`
	// Whether the bug should be updated by Weetbix, and whether failures
	// should still be matched against the rule.
	IsActive bool `json:"isActive"`
	// Whether this rule should manage the priority and verified status
	// of the associated bug based on the impact of the cluster defined
	// by this rule.
	IsManagingBug bool `json:"isManagingBug"`
	// The suggested cluster this rule was created from (if any).
	// Until re-clustering is complete and has reduced the residual impact
	// of the source cluster, this cluster ID tells bug filing to ignore
	// the source cluster when determining whether new bugs need to be filed.
	SourceCluster clustering.ClusterID `json:"sourceCluster"`
}

// Read reads the failure association rule with the given rule ID.
// If no rule exists, NotExistsErr will be returned.
func Read(ctx context.Context, project string, id string) (*FailureAssociationRule, error) {
	whereClause := `Project = @project AND RuleId = @ruleId`
	params := map[string]interface{}{
		"project": project,
		"ruleId":  id,
	}
	rs, err := readWhere(ctx, whereClause, params)
	if err != nil {
		return nil, errors.Annotate(err, "query rule by id").Err()
	}
	if len(rs) == 0 {
		return nil, NotExistsErr
	}
	return rs[0], nil
}

// ReadAll reads all Weetbix failure association rules in a given project.
// This method is not expected to scale -- for testing use only.
func ReadAll(ctx context.Context, project string) ([]*FailureAssociationRule, error) {
	whereClause := `Project = @project`
	params := map[string]interface{}{
		"project": project,
	}
	rs, err := readWhere(ctx, whereClause, params)
	if err != nil {
		return nil, errors.Annotate(err, "query all rules").Err()
	}
	return rs, nil
}

// ReadActive reads all active Weetbix failure association rules in the given LUCI project.
func ReadActive(ctx context.Context, project string) ([]*FailureAssociationRule, error) {
	whereClause := `Project = @project AND IsActive`
	params := map[string]interface{}{
		"project": project,
	}
	rs, err := readWhere(ctx, whereClause, params)
	if err != nil {
		return nil, errors.Annotate(err, "query active rules").Err()
	}
	return rs, nil
}

// ReadByBug reads the failure association rules associated with the given bug.
// At most one rule will be returned per project.
func ReadByBug(ctx context.Context, bugID bugs.BugID) ([]*FailureAssociationRule, error) {
	whereClause := `BugSystem = @bugSystem and BugId = @bugId`
	params := map[string]interface{}{
		"bugSystem": bugID.System,
		"bugId":     bugID.ID,
	}
	rs, err := readWhere(ctx, whereClause, params)
	if err != nil {
		return nil, errors.Annotate(err, "query rule by bug").Err()
	}
	return rs, nil
}

// ReadDelta reads the changed failure association rules since the given
// timestamp, in the given LUCI project.
func ReadDelta(ctx context.Context, project string, sinceTime time.Time) ([]*FailureAssociationRule, error) {
	if sinceTime.Before(StartingEpoch) {
		return nil, errors.New("cannot query rule deltas from before project inception")
	}
	whereClause := `Project = @project AND LastUpdated > @sinceTime`
	params := map[string]interface{}{
		"project":   project,
		"sinceTime": sinceTime,
	}
	rs, err := readWhere(ctx, whereClause, params)
	if err != nil {
		return nil, errors.Annotate(err, "query rules since").Err()
	}
	return rs, nil
}

// ReadMany reads the failure association rules with the given rule IDs.
// The returned slice of rules will correspond one-to-one the IDs requested
// (so returned[i].RuleId == ids[i], assuming the rule exists, else
// returned[i] == nil). If a rule does not exist, a value of nil will be
// returned for that ID. The same rule can be requested multiple times.
func ReadMany(ctx context.Context, project string, ids []string) ([]*FailureAssociationRule, error) {
	whereClause := `Project = @project AND RuleId IN UNNEST(@ruleIds)`
	params := map[string]interface{}{
		"project": project,
		"ruleIds": ids,
	}
	rs, err := readWhere(ctx, whereClause, params)
	if err != nil {
		return nil, errors.Annotate(err, "query rules by id").Err()
	}
	ruleByID := make(map[string]FailureAssociationRule)
	for _, r := range rs {
		ruleByID[r.RuleID] = *r
	}
	var result []*FailureAssociationRule
	for _, id := range ids {
		var entry *FailureAssociationRule
		rule, ok := ruleByID[id]
		if ok {
			// Copy the rule to ensure the rules in the result
			// are not aliased, even if the same rule ID is requested
			// multiple times.
			entry = new(FailureAssociationRule)
			*entry = rule
		}
		result = append(result, entry)
	}
	return result, nil
}

// readWhere failure association rules matching the given where clause,
// substituting params for any SQL parameters used in that clause.
func readWhere(ctx context.Context, whereClause string, params map[string]interface{}) ([]*FailureAssociationRule, error) {
	stmt := spanner.NewStatement(`
		SELECT Project, RuleId, RuleDefinition, BugSystem, BugId,
		  CreationTime, LastUpdated, PredicateLastUpdated,
		  CreationUser, LastUpdatedUser,
		  IsActive, IsManagingBug,
		  SourceClusterAlgorithm, SourceClusterId
		FROM FailureAssociationRules
		WHERE (` + whereClause + `)
		ORDER BY BugSystem, BugId, Project
	`)
	stmt.Params = params

	it := span.Query(ctx, stmt)
	rs := []*FailureAssociationRule{}
	err := it.Do(func(r *spanner.Row) error {
		var project, ruleID, ruleDefinition, bugSystem, bugID string
		var creationTime, lastUpdated, predicateLastUpdated time.Time
		var creationUser, lastUpdatedUser string
		var isActive, isManagingBug spanner.NullBool
		var sourceClusterAlgorithm, sourceClusterID string
		err := r.Columns(
			&project, &ruleID, &ruleDefinition, &bugSystem, &bugID,
			&creationTime, &lastUpdated, &predicateLastUpdated,
			&creationUser, &lastUpdatedUser,
			&isActive, &isManagingBug,
			&sourceClusterAlgorithm, &sourceClusterID,
		)
		if err != nil {
			return errors.Annotate(err, "read rule row").Err()
		}

		rule := &FailureAssociationRule{
			Project:              project,
			RuleID:               ruleID,
			RuleDefinition:       ruleDefinition,
			CreationTime:         creationTime,
			CreationUser:         creationUser,
			LastUpdated:          lastUpdated,
			LastUpdatedUser:      lastUpdatedUser,
			PredicateLastUpdated: predicateLastUpdated,
			BugID:                bugs.BugID{System: bugSystem, ID: bugID},
			IsActive:             isActive.Valid && isActive.Bool,
			IsManagingBug:        isManagingBug.Valid && isManagingBug.Bool,
			SourceCluster: clustering.ClusterID{
				Algorithm: sourceClusterAlgorithm,
				ID:        sourceClusterID,
			},
		}
		rs = append(rs, rule)
		return nil
	})
	return rs, err
}

// Version captures version information about a project's rules.
type Version struct {
	// Predicates is the last time any rule changed its
	// rule predicate (RuleDefinition or IsActive).
	// Also known as "Rules Version" in clustering contexts.
	Predicates time.Time
	// Total is the last time any rule was updated in any way.
	// Pass to ReadDelta when seeking to read changed rules.
	Total time.Time
}

// ReadVersion reads information about when rules in the given project
// were last updated. This is used to version the set of rules retrieved
// by ReadActive and is typically called in the same transaction.
// It is also used to implement change detection on rule predicates
// for the purpose of triggering re-clustering.
//
// Simply reading the last LastUpdated time of the rules read by ReadActive
// is not sufficient to version the set of rules read, as the most recent
// update may have been to mark a rule inactive (removing it from the set
// that is read).
//
// If the project has no failure association rules, the timestamp
// StartingEpoch is returned.
func ReadVersion(ctx context.Context, projectID string) (Version, error) {
	stmt := spanner.NewStatement(`
		SELECT
		  Max(PredicateLastUpdated) as PredicateLastUpdated,
		  MAX(LastUpdated) as LastUpdated
		FROM FailureAssociationRules
		WHERE Project = @projectID
	`)
	stmt.Params = map[string]interface{}{
		"projectID": projectID,
	}
	var predicateLastUpdated, lastUpdated spanner.NullTime
	it := span.Query(ctx, stmt)
	err := it.Do(func(r *spanner.Row) error {
		err := r.Columns(&predicateLastUpdated, &lastUpdated)
		if err != nil {
			return errors.Annotate(err, "read last updated row").Err()
		}
		return nil
	})
	if err != nil {
		return Version{}, errors.Annotate(err, "query last updated").Err()
	}
	result := Version{
		Predicates: StartingEpoch,
		Total:      StartingEpoch,
	}
	// predicateLastUpdated / lastUpdated are only invalid if there
	// are no failure association rules.
	if predicateLastUpdated.Valid {
		result.Predicates = predicateLastUpdated.Time
	}
	if lastUpdated.Valid {
		result.Total = lastUpdated.Time
	}
	return result, nil
}

// ReadTotalActiveRules reads the number active rules, for each LUCI Project.
// Only returns entries for projects that have any rules (at all). Combine
// with config if you need zero entries for projects that are defined but
// have no rules.
func ReadTotalActiveRules(ctx context.Context) (map[string]int64, error) {
	stmt := spanner.NewStatement(`
		SELECT
		  project,
		  COUNTIF(IsActive) as active_rules,
		FROM FailureAssociationRules
		GROUP BY project
	`)
	result := make(map[string]int64)
	it := span.Query(ctx, stmt)
	err := it.Do(func(r *spanner.Row) error {
		var project string
		var activeRules int64
		err := r.Columns(&project, &activeRules)
		if err != nil {
			return errors.Annotate(err, "read row").Err()
		}
		result[project] = activeRules
		return nil
	})
	if err != nil {
		return nil, errors.Annotate(err, "query total active rules by project").Err()
	}
	return result, nil
}

// Create inserts a new failure association rule with the specified details.
func Create(ctx context.Context, r *FailureAssociationRule, user string) error {
	if err := validateRule(r); err != nil {
		return err
	}
	if err := validateUser(user); err != nil {
		return err
	}
	ms := spanutil.InsertMap("FailureAssociationRules", map[string]interface{}{
		"Project":              r.Project,
		"RuleId":               r.RuleID,
		"RuleDefinition":       r.RuleDefinition,
		"PredicateLastUpdated": spanner.CommitTimestamp,
		"CreationTime":         spanner.CommitTimestamp,
		"CreationUser":         user,
		"LastUpdated":          spanner.CommitTimestamp,
		"LastUpdatedUser":      user,
		"BugSystem":            r.BugID.System,
		"BugId":                r.BugID.ID,
		// IsActive uses the value 'NULL' to indicate false, and true to indicate true.
		"IsActive":               spanner.NullBool{Bool: r.IsActive, Valid: r.IsActive},
		"IsManagingBug":          r.IsManagingBug,
		"SourceClusterAlgorithm": r.SourceCluster.Algorithm,
		"SourceClusterId":        r.SourceCluster.ID,
	})
	span.BufferWrite(ctx, ms)
	return nil
}

// Update updates an existing failure association rule to have the specified
// details. Set updatePredicate to true if you changed RuleDefinition
// or IsActive.
func Update(ctx context.Context, r *FailureAssociationRule, updatePredicate bool, user string) error {
	if err := validateRule(r); err != nil {
		return err
	}
	if err := validateUser(user); err != nil {
		return err
	}
	update := map[string]interface{}{
		"Project":                r.Project,
		"RuleId":                 r.RuleID,
		"LastUpdated":            spanner.CommitTimestamp,
		"LastUpdatedUser":        user,
		"BugSystem":              r.BugID.System,
		"BugId":                  r.BugID.ID,
		"SourceClusterAlgorithm": r.SourceCluster.Algorithm,
		"SourceClusterId":        r.SourceCluster.ID,
		"IsManagingBug":          r.IsManagingBug,
	}
	if updatePredicate {
		update["RuleDefinition"] = r.RuleDefinition
		// IsActive uses the value 'NULL' to indicate false, and true to indicate true.
		update["IsActive"] = spanner.NullBool{Bool: r.IsActive, Valid: r.IsActive}
		update["PredicateLastUpdated"] = spanner.CommitTimestamp
	}
	ms := spanutil.UpdateMap("FailureAssociationRules", update)
	span.BufferWrite(ctx, ms)
	return nil
}

func validateRule(r *FailureAssociationRule) error {
	switch {
	case !config.ProjectRe.MatchString(r.Project):
		return errors.New("project must be valid")
	case !RuleIDRe.MatchString(r.RuleID):
		return errors.New("rule ID must be valid")
	case r.BugID.Validate() != nil:
		return errors.Annotate(r.BugID.Validate(), "bug ID is not valid").Err()
	case r.SourceCluster.Validate() != nil && !r.SourceCluster.IsEmpty():
		return errors.Annotate(r.SourceCluster.Validate(), "source cluster ID is not valid").Err()
	}
	_, err := lang.Parse(r.RuleDefinition)
	if err != nil {
		return errors.Annotate(err, "rule definition is not valid").Err()
	}
	return nil
}

func validateUser(u string) error {
	if !UserRe.MatchString(u) {
		return errors.New("user must be valid")
	}
	return nil
}

// GenerateID returns a random 128-bit rule ID, encoded as
// 32 lowercase hexadecimal characters.
func GenerateID() (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}
