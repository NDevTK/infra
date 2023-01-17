// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	pb "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	bbpb "go.chromium.org/luci/buildbucket/proto"
)

type Rule struct {
	rule           *pb.RetryRule
	totalRetries   uint32
	retriesByBuild map[string]uint32
}

// CollectState tracks state for a conductor collect.
type CollectState struct {
	rules []*Rule
}

// initCollectState returns a new CollectState based on the specified config.
func initCollectState(config *pb.CollectConfig) *CollectState {
	rules := []*Rule{}
	for _, rule := range config.GetRules() {
		rules = append(rules, &Rule{
			rule:           rule,
			totalRetries:   0,
			retriesByBuild: map[string]uint32{},
		})
	}
	return &CollectState{
		rules: rules,
	}
}

func (r *Rule) matches(build *bbpb.Build) bool {
	// TODO(b/264680777): Implement.
	return true
}

// canRetry evaluates whether a retry is permitted by all matching rules.
func (c *CollectState) canRetry(build *bbpb.Build) bool {
	buildName := build.GetBuilder().GetBuilder()
	for _, rule := range c.rules {
		if !rule.matches(build) {
			continue
		}
		if rule.rule.GetMaxRetries() > 0 {
			if rule.totalRetries >= uint32(rule.rule.GetMaxRetries()) {
				return false
			}
		}
		if rule.rule.GetMaxRetriesPerBuild() > 0 {
			buildRetries, ok := rule.retriesByBuild[buildName]
			if ok && buildRetries >= uint32(rule.rule.GetMaxRetriesPerBuild()) {
				return false
			}
		}
	}
	// No retries if there are no rules configured.
	return len(c.rules) > 0
}

// recordRetry records that the build was retried.
func (c *CollectState) recordRetry(build *bbpb.Build) {
	buildName := build.GetBuilder().GetBuilder()
	for _, rule := range c.rules {
		if rule.matches(build) {
			rule.totalRetries += 1
			if _, ok := rule.retriesByBuild[buildName]; ok {
				rule.retriesByBuild[buildName] += 1
			} else {
				rule.retriesByBuild[buildName] = 1
			}
		}
	}
}
