// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bugclusters

import (
	"context"
	"regexp"
	"time"

	"cloud.google.com/go/spanner"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/config"
	spanutil "infra/appengine/weetbix/internal/span"
)

// ClusterIDRe matches validly formed cluster IDs.
var ClusterIDRe = regexp.MustCompile(`^[0-9a-f]{32}$`)

// FailureAssociationRules associate failures with bugs. When the rules
// are used to match incoming test failures, the resultant clusters are
// known as 'bug clusters' because clusters are associated with a bug
// (via the failure association rule.
type FailureAssociationRule struct {
	// The LUCI Project for which this bug is being tracked.
	Project string `json:"project"`
	// The unique identifier for the failure association rule,
	// as 32 lowercase hexadecimal characters.
	RuleID string `json:"clusterId"`
	// The rule predicate, defining which failures are being associated.
	RuleDefinition string `json:"ruleDefinition"`
	// The time the rule was created. Output only.
	CreationTime time.Time
	// The time the rule was last updated. Output only.
	LastUpdated time.Time
	// Bug is the identifier of the bug. For monorail, the scheme is
	// monorail/{monorail_project}/{numeric_id}.
	Bug string `json:"bug"`
	// Whether the bug should be updated by Weetbix, and whether failures
	// should still be matched against the rule.
	IsActive bool `json:"isActive"`
	// The suggested cluster this bug cluster was created from (if any).
	// Until re-clustering is complete and has reduced the residual impact
	// of the source cluster, this cluster ID tells bug filing to ignore
	// the source cluster when determining whether new bugs need to be filed.
	SourceCluster clustering.ClusterID
}

// ReadActive reads all active Weetbix failure association rules in the given LUCI project.
func ReadActive(ctx context.Context, projectID string) ([]*FailureAssociationRule, error) {
	stmt := spanner.NewStatement(`
		SELECT Project, RuleId, RuleDefinition, Bug,
		  CreationTime, LastUpdated,
		  SourceClusterAlgorithm, SourceClusterId
		FROM BugClusters
		WHERE IsActive AND Project = @projectID
		ORDER BY Project, Bug
	`)
	stmt.Params = map[string]interface{}{
		"projectID": projectID,
	}
	it := span.Query(ctx, stmt)
	bcs := []*FailureAssociationRule{}
	err := it.Do(func(r *spanner.Row) error {
		var project, ruleID, ruleDefinition, bug string
		var creationTime, lastUpdated time.Time
		var sourceClusterAlgorithm, sourceClusterID string
		err := r.Columns(
			&project, &ruleID, &ruleDefinition, &bug,
			&creationTime, &lastUpdated,
			&sourceClusterAlgorithm, &sourceClusterID,
		)
		if err != nil {
			return errors.Annotate(err, "read bug cluster row").Err()
		}

		bc := &FailureAssociationRule{
			Project:        project,
			RuleID:         ruleID,
			RuleDefinition: ruleDefinition,
			CreationTime:   creationTime,
			LastUpdated:    lastUpdated,
			Bug:            bug,
			IsActive:       true,
			SourceCluster: clustering.ClusterID{
				Algorithm: sourceClusterAlgorithm,
				ID:        sourceClusterID,
			},
		}
		bcs = append(bcs, bc)
		return nil
	})
	if err != nil {
		return nil, errors.Annotate(err, "query active bug clusters").Err()
	}
	return bcs, nil
}

// Create inserts a new failure assoication rule with the specified details.
func Create(ctx context.Context, r *FailureAssociationRule) error {
	if err := validateRule(r); err != nil {
		return err
	}
	ms := spanutil.InsertMap("BugClusters", map[string]interface{}{
		"Project":        r.Project,
		"RuleId":         r.RuleID,
		"RuleDefinition": r.RuleDefinition,
		"CreationTime":   spanner.CommitTimestamp,
		"LastUpdated":    spanner.CommitTimestamp,
		"Bug":            r.Bug,
		// IsActive uses the value 'NULL' to indicate false, and true to indicate true.
		"IsActive":               spanner.NullBool{Bool: r.IsActive, Valid: r.IsActive},
		"SourceClusterAlgorithm": r.SourceCluster.Algorithm,
		"SourceClusterId":        r.SourceCluster.ID,
	})
	span.BufferWrite(ctx, ms)
	return nil
}

func validateRule(r *FailureAssociationRule) error {
	switch {
	case !config.ProjectRe.MatchString(r.Project):
		return errors.New("project must be valid")
	// The Rule ID is also used as the cluster ID by the failure association
	// rule-based clustering algorithm.
	case !ClusterIDRe.MatchString(r.RuleID):
		return errors.New("rule ID must be valid")
	case r.Bug == "":
		return errors.New("bug must be specified")
	case r.SourceCluster.Algorithm == "" || clustering.AlgorithmRe.MatchString(r.SourceCluster.Algorithm):
		return errors.New("source cluster algorithm must be empty or valid")
	case r.SourceCluster.Validate() != nil && !r.SourceCluster.IsEmpty():
		return errors.Annotate(r.SourceCluster.Validate(), "source cluster ID is not valid").Err()
	}
	return nil
}
