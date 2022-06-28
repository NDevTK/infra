// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"fmt"
	"regexp"

	"go.chromium.org/luci/common/errors"

	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/clustering/rules"
	"infra/appengine/weetbix/internal/config"
)

// Regular expressions for matching resource names used in APIs.
var (
	GenericKeyPattern = "[a-z0-9\\-]+"
	RuleNameRe        = regexp.MustCompile(`^projects/(` + config.ProjectRePattern + `)/rules/(` + rules.RuleIDRePattern + `)$`)
	// ClusterNameRe performs partial validation of a
	// cluster resource name. Cluster algorithm and
	// ID must be further validated by ClusterID.Validate().
	ClusterNameRe = regexp.MustCompile(`^projects/(` + config.ProjectRePattern + `)/clusters/(` + GenericKeyPattern + `)/(` + GenericKeyPattern + `)$`)
	ProjectNameRe = regexp.MustCompile(`^projects/(` + config.ProjectRePattern + `)$`)
)

// parseRuleName parses a rule resource name into its constituent ID parts.
func parseRuleName(name string) (project, ruleID string, err error) {
	match := RuleNameRe.FindStringSubmatch(name)
	if match == nil {
		return "", "", errors.New("invalid rule name, expected format: projects/{project}/rules/{rule_id}")
	}
	return match[1], match[2], nil
}

// parseProjectName parses a project resource name into a project ID.
func parseProjectName(name string) (project string, err error) {
	match := ProjectNameRe.FindStringSubmatch(name)
	if match == nil {
		return "", errors.New("invalid project name, expected format: projects/{project}")
	}
	return match[1], nil
}

// parseClusterName parses a cluster resource name into its constituent ID
// parts. Algorithm aliases are resolved to concrete algorithm names.
func parseClusterName(name string) (project string, clusterID clustering.ClusterID, err error) {
	match := ClusterNameRe.FindStringSubmatch(name)
	if match == nil {
		return "", clustering.ClusterID{}, errors.New("invalid cluster name, expected format: projects/{project}/clusters/{cluster_alg}/{cluster_id}")
	}
	algorithm := resolveAlgorithm(match[2])
	id := match[3]
	cID := clustering.ClusterID{Algorithm: algorithm, ID: id}
	if err := cID.Validate(); err != nil {
		return "", clustering.ClusterID{}, errors.Annotate(err, "invalid cluster name").Err()
	}
	return match[1], cID, nil
}

func ruleName(project, ruleID string) string {
	return fmt.Sprintf("projects/%s/rules/%s", project, ruleID)
}
