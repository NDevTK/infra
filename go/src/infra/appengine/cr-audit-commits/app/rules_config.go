// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package crauditcommits implements cr-audit-commits.appspot.com services.
package crauditcommits

import (
	"go.chromium.org/luci/common/api/gitiles"
)

// RuleMap maps each monitored repository to a list of account/rules structs.
var RuleMap = map[string][]RuleSet{
	"chromium-src-master": {
		AccountRules{
			Account: "findit-for-me@appspot.gserviceaccount.com",
			Funcs:   []RuleFunc{DummyRule},
		},
	},
}

// RuleSet is a group of rules that can be decided to apply or not to a
// specific commit, as a unit.
type RuleSet interface {
	Matches(gitiles.Commit) bool
}

// AccountRules is a RuleSet that applies to a commit if the commit has a given
// account as either its author or its committer.
type AccountRules struct {
	Account string
	Funcs   []RuleFunc
}

// Matches determines whether the AccountRules set it's bound to applies to the
// given commit.
func (ar AccountRules) Matches(c gitiles.Commit) bool {
	if c.Committer.Email == ar.Account || c.Author.Email == ar.Account {
		return true
	}
	return false
}

// RuleFunc is the function type for audit rules.
type RuleFunc func(*gitiles.Commit) *RuleResult

// DummyRule is useless, only meant as a placeholder.
//
// Rules are expected to accept a Commit and return a
// reference to a RuleResult
func DummyRule(c *gitiles.Commit) *RuleResult {
	result := &RuleResult{}
	result.RuleName = "dummy_rule_name"
	if c.Commit != "deadbeef" {
		result.RuleResultStatus = rulePassed
	} else {
		result.RuleResultStatus = ruleFailed
		result.Message = "DummyProperty or the commit hash was deadbeef"
	}
	return result
}
