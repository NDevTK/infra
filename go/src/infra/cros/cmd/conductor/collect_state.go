// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	pb "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	bbpb "go.chromium.org/luci/buildbucket/proto"
)

type Clock interface {
	Now() int64
}

type RealClock struct{}

func (c *RealClock) Now() int64 {
	return time.Now().Unix()
}

type Rule struct {
	rule           *pb.RetryRule
	totalRetries   uint32
	retriesByBuild map[string]uint32
}

// CollectState tracks state for a conductor collect.
type CollectState struct {
	stdoutLog *log.Logger
	stderrLog *log.Logger

	rules     []*Rule
	startTime int64
	clock     Clock
}

// LogOut logs to stdout.
func (c *CollectState) LogOut(format string, a ...interface{}) {
	if c.stdoutLog != nil {
		c.stdoutLog.Printf(format, a...)
	}
}

// LogErr logs to stderr.
func (c *CollectState) LogErr(format string, a ...interface{}) {
	if c.stderrLog != nil {
		c.stderrLog.Printf(format, a...)
	}
}

// initCollectState returns a new CollectState based on the specified config.
func initCollectState(config *pb.CollectConfig, stdoutLog, stderrLog *log.Logger) *CollectState {
	clock := &RealClock{}
	return &CollectState{
		clock:     clock,
		stdoutLog: stdoutLog,
		stderrLog: stderrLog,
		rules:     initRules(config),
		startTime: clock.Now(),
	}
}

// initCollectStateTest returns a new CollectState based on the specified
// config that uses the specified clock.
func initCollectStateTest(config *pb.CollectConfig, clock Clock) *CollectState {
	return &CollectState{
		rules:     initRules(config),
		clock:     clock,
		startTime: clock.Now(),
	}
}

func initRules(config *pb.CollectConfig) []*Rule {
	rules := []*Rule{}
	for _, rule := range config.GetRules() {
		rules = append(rules, &Rule{
			rule:           rule,
			totalRetries:   0,
			retriesByBuild: map[string]uint32{},
		})
	}
	return rules
}

func (r *Rule) matches(build *bbpb.Build) bool {
	if len(r.rule.GetStatus()) > 0 {
		status := build.GetStatus()
		statusMatch := false
		for _, ruleStatus := range r.rule.GetStatus() {
			if status == bbpb.Status(ruleStatus) {
				statusMatch = true
				break
			}
		}
		if !statusMatch {
			return false
		}
	}
	return true
}

// canRetry evaluates whether a retry is permitted by all matching rules.
func (c *CollectState) canRetry(build *bbpb.Build) bool {
	buildName := build.GetBuilder().GetBuilder()
	currentTime := c.clock.Now()
	matchesRules := []string{}
	for i, rule := range c.rules {
		if !rule.matches(build) {
			continue
		}
		matchesRules = append(matchesRules, fmt.Sprintf("%d", i))
	}
	buildStr := fmt.Sprintf("(%s, %d, %s)", buildName, build.GetId(), build.GetStatus())
	if len(matchesRules) == 0 {
		c.LogOut("Build %s does not match any rules, not evaluating for retry", buildStr)
		return false
	}

	c.LogOut("Build %s matches rules %s, evaluating for retry", buildStr, strings.Join(matchesRules, ","))
	for i, rule := range c.rules {
		if !rule.matches(build) {
			continue
		}
		if rule.rule.GetMaxRetries() > 0 {
			if rule.totalRetries >= uint32(rule.rule.GetMaxRetries()) {
				c.LogOut("Rule %d has used %d/%d total retries, not retrying.",
					i, rule.totalRetries, rule.rule.GetMaxRetries())
				return false
			}
		}
		if rule.rule.GetMaxRetriesPerBuild() > 0 {
			buildRetries, ok := rule.retriesByBuild[buildName]
			if ok && buildRetries >= uint32(rule.rule.GetMaxRetriesPerBuild()) {
				c.LogOut("Rule %d has used %d/%d total retries for build %s, not retrying.",
					i, buildRetries, rule.rule.GetMaxRetriesPerBuild(), buildName)
				return false
			}
		}
		if rule.rule.GetCutoffSeconds() > 0 {
			if c.startTime+int64(rule.rule.GetCutoffSeconds()) < currentTime {
				c.LogOut("Rule %d only allows retries %d seconds into the collection"+
					" (we're at %d seconds), not retrying.",
					i, rule.rule.GetCutoffSeconds(), currentTime-c.startTime)
				return false
			}
		}
	}
	return true
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
