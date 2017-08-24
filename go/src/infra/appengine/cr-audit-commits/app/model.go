// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package crauditcommits implements cr-audit-commits.appspot.com services.
package crauditcommits

import (
	"go.chromium.org/gae/service/datastore"
)

// AuditStatus is the enum for RelevantCommit.Status.
type AuditStatus int

const (
	auditScheduled AuditStatus = iota
	auditCompleted
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
	RepoURL            string `gae:"$id"`
	Name               string
	LastKnownCommit    string
	LastRelevantCommit string
}

// RelevantCommit points to a node in a linked list of commits that have
// been considered relevant by CommitScanner.
type RelevantCommit struct {
	ForRepoConfig          *datastore.Key `gae:"$parent"`
	CommitHash             string         `gae:"$id"`
	PreviousRelevantCommit string
	Status                 AuditStatus
	Result                 []*RuleResult
}

// RuleResult represents the result of applying a single rule to a commit.
type RuleResult struct {
	RuleName         string
	RuleResultStatus RuleStatus
	Message          string
}
