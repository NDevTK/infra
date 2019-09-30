// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"strings"

	"go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"

	"infra/cmd/cros_test_platform/internal/site"

	"github.com/golang/protobuf/proto"
	"github.com/maruel/subcommands"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/config"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/migration/scheduler"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/steps"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	gitilespb "go.chromium.org/luci/common/proto/gitiles"
)

// SchedulerTrafficSplit implements the `scheduler-traffic-split` subcommand.
var SchedulerTrafficSplit = &subcommands.Command{
	UsageLine: "scheduler-traffic-split -input_json /path/to/input.json -output_json /path/to/output.json",
	ShortDesc: "Determine traffic split between backend schedulers.",
	LongDesc: `Determine traffic split between backend schedulers, i.e. Autotest vs Skylab.

Step input and output is JSON encoded protobuf defined at
https://chromium.googlesource.com/chromiumos/infra/proto/+/master/src/test_platform/steps/scheduler_traffic_split.proto`,
	CommandRun: func() subcommands.CommandRun {
		c := &schedulerTrafficSplitRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.Flags.StringVar(&c.inputPath, "input_json", "", "Path that contains JSON encoded test_platform.steps.SchedulerTrafficSplitRequest")
		c.Flags.StringVar(&c.outputPath, "output_json", "", "Path where JSON encoded test_platform.steps.SchedulerTrafficSplitResponse should be written.")
		return c
	},
}

type schedulerTrafficSplitRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags

	inputPath  string
	outputPath string
}

func (c *schedulerTrafficSplitRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	err := c.innerRun(a, args, env)
	if err != nil {
		fmt.Fprintf(a.GetErr(), "%s\n", err)
	}
	return exitCode(err)
}

func (c *schedulerTrafficSplitRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.processCLIArgs(args); err != nil {
		return err
	}
	var request steps.SchedulerTrafficSplitRequest
	if err := readRequest(c.inputPath, &request); err != nil {
		return err
	}

	ctx := cli.GetContext(a, c, env)
	ctx = setupLogging(ctx)

	split, err := c.getTrafficSplitConfig(ctx, request.Config)
	if err != nil {
		return err
	}
	resp, err := determineTrafficSplit(request.Request, split)
	if err != nil {
		logPotentiallyRelevantRules(ctx, request.Request, split.Rules)
		return err
	}
	return writeResponse(c.outputPath, resp, nil)
}

func (c *schedulerTrafficSplitRun) processCLIArgs(args []string) error {
	if len(args) > 0 {
		return errors.Reason("have %d positional args, want 0", len(args)).Err()
	}
	if c.inputPath == "" {
		return errors.Reason("-input_json not specified").Err()
	}
	if c.outputPath == "" {
		return errors.Reason("-output_json not specified").Err()
	}
	return nil
}

func (c *schedulerTrafficSplitRun) getTrafficSplitConfig(ctx context.Context, config *config.Config_SchedulerMigration) (*scheduler.TrafficSplit, error) {
	g, err := c.newGitilesClient(ctx, config.GitilesHost)
	if err != nil {
		return nil, errors.Annotate(err, "get traffic split config").Err()
	}
	text, err := c.downloadTrafficSplitConfig(ctx, g, config)
	if err != nil {
		return nil, errors.Annotate(err, "get traffic split config").Err()
	}
	var split scheduler.TrafficSplit
	if err := unmarshaller.Unmarshal(strings.NewReader(text), &split); err != nil {
		return nil, errors.Annotate(err, "get traffic split config").Err()
	}
	return &split, nil
}

func (c *schedulerTrafficSplitRun) newGitilesClient(ctx context.Context, host string) (gitilespb.GitilesClient, error) {
	h, err := newAuthenticatedHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return nil, errors.Annotate(err, "new gitiles client").Err()
	}
	return gitiles.NewRESTClient(h, host, true)
}

// downloadTrafficSplitConfig returns the contents of the config downloaded from Gitiles.
func (c *schedulerTrafficSplitRun) downloadTrafficSplitConfig(ctx context.Context, client gitilespb.GitilesClient, config *config.Config_SchedulerMigration) (string, error) {
	res, err := client.DownloadFile(ctx, &gitilespb.DownloadFileRequest{
		Project:    config.GitProject,
		Committish: config.Commitish,
		Path:       config.FilePath,
		Format:     gitilespb.DownloadFileRequest_TEXT,
	})
	if err != nil {
		return "", errors.Annotate(err, "download from gitiles").Err()
	}
	return res.Contents, nil
}

func determineTrafficSplit(request *test_platform.Request, trafficSplitConfig *scheduler.TrafficSplit) (*steps.SchedulerTrafficSplitResponse, error) {
	if err := ensureSufficientForTrafficSplit(request); err != nil {
		return nil, errors.Annotate(err, "determine traffic split").Err()
	}

	rules := determineRelevantSuiteRules(request, trafficSplitConfig.SuiteOverrides)
	if len(rules) == 0 {
		rules = newRuleFilter(trafficSplitConfig.Rules).ForRequest(request)
	}

	var rule *scheduler.Rule
	switch {
	case len(rules) == 0:
		return nil, errors.Reason("no matching traffic split rule").Err()
	case len(rules) == 1:
		rule = rules[0]
	default:
		if err := ensureRulesAreCompatible(rules); err != nil {
			return nil, errors.Annotate(err, "determine traffic split").Err()
		}
		rule = rules[0]
	}
	return applyTrafficSplitRule(request, rule)
}

func ensureSufficientForTrafficSplit(r *test_platform.Request) error {
	if r.GetParams().GetScheduling().GetPool() == nil {
		return errors.Reason("request contains no pool information").Err()
	}
	return nil
}

func ensureRulesAreCompatible(rules []*scheduler.Rule) error {
	b := rules[0].GetBackend()
	s := rules[0].GetRequestMod().GetScheduling()
	for _, r := range rules[1:] {
		if r.GetBackend() != b {
			return errors.Reason("Rules %s and %s contain conflicting backends", rules[0], r).Err()
		}
		if schedulingNotEqual(s, r.GetRequestMod().GetScheduling()) {
			return errors.Reason("Rules %s and %s contain conflicting request modifications", rules[0], r).Err()
		}
	}
	return nil
}

func schedulingNotEqual(s1, s2 *test_platform.Request_Params_Scheduling) bool {
	if s1.GetUnmanagedPool() != s2.GetUnmanagedPool() {
		return true
	}
	if s1.GetManagedPool() != s2.GetManagedPool() {
		return true
	}
	if s1.GetQuotaAccount() != s2.GetQuotaAccount() {
		return true
	}
	return false
}

func applyTrafficSplitRule(request *test_platform.Request, rule *scheduler.Rule) (*steps.SchedulerTrafficSplitResponse, error) {
	newRequest := applyRequestModification(request, rule.GetRequestMod())
	switch rule.Backend {
	case scheduler.Backend_BACKEND_AUTOTEST:
		return &steps.SchedulerTrafficSplitResponse{
			AutotestRequest: newRequest,
		}, nil
	case scheduler.Backend_BACKEND_SKYLAB:
		return &steps.SchedulerTrafficSplitResponse{
			SkylabRequest: newRequest,
		}, nil
	default:
		return nil, errors.Reason("invalid backend %s in rule", rule.Backend.String()).Err()
	}
}

func applyRequestModification(request *test_platform.Request, mod *scheduler.RequestMod) *test_platform.Request {
	if mod == nil {
		return request
	}
	var dst test_platform.Request
	proto.Merge(&dst, request)
	if dst.Params == nil {
		dst.Params = &test_platform.Request_Params{}
	}
	proto.Merge(dst.Params.Scheduling, mod.Scheduling)
	return &dst
}

func determineRelevantSuiteRules(request *test_platform.Request, suiteOverrides []*scheduler.SuiteOverride) []*scheduler.Rule {
	suites := make(stringset.Set)
	for _, s := range request.GetTestPlan().GetSuite() {
		if s.GetName() != "" {
			suites.Add(s.GetName())
		}
	}

	rules := []*scheduler.Rule{}
	for _, so := range suiteOverrides {
		r := so.Rule
		if suites.Has(so.GetSuite().GetName()) && isRuleRelevant(request, r) {
			rules = append(rules, r)
		}
	}
	return rules
}

func logPotentiallyRelevantRules(ctx context.Context, request *test_platform.Request, rules []*scheduler.Rule) {
	f := newRuleFilter(rules)
	logger := logging.Get(ctx)
	logger.Warningf("No matching rule found. Printing partially matching rules...")

	m := request.GetParams().GetHardwareAttributes().GetModel()
	if pr := f.ForModel(m); len(pr) > 0 {
		logger.Infof("Following rules match requested model: %s", formatFirstFewRules(pr))
	} else {
		logger.Warningf("No rules matched requested model %s.", m)
	}

	b := request.GetParams().GetSoftwareAttributes().GetBuildTarget().GetName()
	if pr := f.ForBuildTarget(b); len(pr) > 0 {
		logger.Infof("Following rules match requested buildTarget: %s", formatFirstFewRules(pr))
	} else {
		logger.Warningf("No rules matched requested build target %s.", b)
	}

	s := request.GetParams().GetScheduling()
	if pr := f.ForScheduling(s); len(pr) > 0 {
		logger.Infof("Following rules match requested scheduling: %s", formatFirstFewRules(pr))
	} else {
		logger.Warningf("No rules matched requested scheduling %s.", s)
	}
}

func formatFirstFewRules(rules []*scheduler.Rule) string {
	const rulesToPrint = 5
	s := fmt.Sprintf("%v", rules[:rulesToPrint])
	if len(s) > rulesToPrint {
		s = fmt.Sprintf("%s... [%d more]", s, len(s)-rulesToPrint)
	}
	return s
}

type ruleFilter []*scheduler.Rule

func newRuleFilter(rules []*scheduler.Rule) ruleFilter {
	return ruleFilter(rules)
}

// ForRequest returns rules relevant to a test platform request.
func (f ruleFilter) ForRequest(request *test_platform.Request) []*scheduler.Rule {
	ret := []*scheduler.Rule{}
	for _, r := range f {
		if isRuleRelevant(request, r) {
			ret = append(ret, r)
		}
	}
	return ret
}

func isRuleRelevant(request *test_platform.Request, rule *scheduler.Rule) bool {
	if isNonEmptyAndDistinct(
		request.GetParams().GetSoftwareAttributes().GetBuildTarget().GetName(),
		rule.GetRequest().GetBuildTarget().GetName(),
	) {
		return false
	}
	if isNonEmptyAndDistinct(
		request.GetParams().GetHardwareAttributes().GetModel(),
		rule.GetRequest().GetModel(),
	) {
		return false
	}
	return isSchedulingRelevant(request.GetParams().GetScheduling(), rule.GetRequest().GetScheduling())
}

func isSchedulingRelevant(got, want *test_platform.Request_Params_Scheduling) bool {
	if isNonEmptyAndDistinct(got.GetUnmanagedPool(), want.GetUnmanagedPool()) {
		return false
	}
	if isNonEmptyAndDistinct(got.GetManagedPool().String(), want.GetManagedPool().String()) {
		return false
	}
	if isNonEmptyAndDistinct(got.GetQuotaAccount(), want.GetQuotaAccount()) {
		return false
	}
	return true
}

func isNonEmptyAndDistinct(got, want string) bool {
	return got != "" && got != want
}

// ForModel returns rules relevant to a model.
func (f ruleFilter) ForModel(model string) []*scheduler.Rule {
	return f.ForRequest(&test_platform.Request{
		Params: &test_platform.Request_Params{
			HardwareAttributes: &test_platform.Request_Params_HardwareAttributes{
				Model: model,
			},
		},
	})
}

// ForBuildTarget returns rules relevant to a build target.
func (f ruleFilter) ForBuildTarget(buildTarget string) []*scheduler.Rule {
	return f.ForRequest(&test_platform.Request{
		Params: &test_platform.Request_Params{
			SoftwareAttributes: &test_platform.Request_Params_SoftwareAttributes{
				BuildTarget: &chromiumos.BuildTarget{
					Name: buildTarget,
				},
			},
		},
	})
}

// ForScheduling returns rules relevant to a scheduling argument.
func (f ruleFilter) ForScheduling(s *test_platform.Request_Params_Scheduling) []*scheduler.Rule {
	return f.ForRequest(&test_platform.Request{
		Params: &test_platform.Request_Params{
			Scheduling: s,
		},
	})
}
