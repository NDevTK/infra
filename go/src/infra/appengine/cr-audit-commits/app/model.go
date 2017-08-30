// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package crauditcommits implements cr-audit-commits.appspot.com services.
package crauditcommits

import (
	"time"

	ds "go.chromium.org/gae/service/datastore"
)

// AuditStatus is the enum for RelevantCommit.Status.
type AuditStatus int

const (
	auditScheduled AuditStatus = iota
	auditCompleted
	auditCompletedWithViolation
)

// RuleStatus is the enum for RuleResult.RuleResultStatus.
type RuleStatus int

const (
	rulePassed RuleStatus = iota
	ruleFailed
	ruleSkipped
)

// RepoConfig contains the configuration and state for each repository we audit.
type RepoConfig struct {
	Name string `gae:"$id"`

	BaseRepoURL        string
	GerritURL          string
	BranchName         string
	LastKnownCommit    string
	LastRelevantCommit string
	StartingCommit     string
}

// RepoURL composes the url of the repository by appending the branch.
func (rc *RepoConfig) RepoURL() string {
	return rc.BaseRepoURL + "/+/" + rc.BranchName
}

// RelevantCommit points to a node in a linked list of commits that have
// been considered relevant by CommitScanner.
type RelevantCommit struct {
	RepoConfigKey *ds.Key `gae:"$parent"`
	CommitHash    string  `gae:"$id"`

	PreviousRelevantCommit string
	Status                 AuditStatus
	Result                 []*RuleResult
	CommitTime             time.Time
	CommitterAccount       string
	AuthorAccount          string
}

// RuleResult represents the result of applying a single rule to a commit.
type RuleResult struct {
	RuleName         string
	RuleResultStatus RuleStatus
	Message          string
}
