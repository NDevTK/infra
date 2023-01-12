// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	pb "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	bbpb "go.chromium.org/luci/buildbucket/proto"
)

type Rule struct {
	rule *pb.RetryRule
}

// CollectState tracks state for a conductor collect.
type CollectState struct {
	rules        []*Rule
	totalRetries uint32
}

// initCollectState returns a new CollectState based on the specified config.
func initCollectState(config *pb.CollectConfig) *CollectState {
	rules := []*Rule{}
	for _, rule := range config.GetRules() {
		rules = append(rules, &Rule{
			rule: rule,
		})
	}
	return &CollectState{
		rules:        rules,
		totalRetries: 0,
	}
}

func (r *Rule) matches(build *bbpb.Build) bool {
	// TODO(b/264680777): Implement.
	return true
}

// canRetry evaluates whether a retry is permitted by all matching rules.
func (c *CollectState) canRetry(build *bbpb.Build) bool {
	for _, rule := range c.rules {
		if !rule.matches(build) {
			continue
		}
		if rule.rule.GetMaxRetries() > 0 {
			if c.totalRetries >= uint32(rule.rule.GetMaxRetries()) {
				return false
			}
		}
	}
	return true
}

// recordRetry records that the build was retried.
func (c *CollectState) recordRetry(build *bbpb.Build) {
	c.totalRetries += 1
}
