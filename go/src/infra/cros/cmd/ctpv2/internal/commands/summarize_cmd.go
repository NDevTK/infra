// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/ctpv2/data"
)

// SummarizeCmd represents summarizing of all test results cmd.
type SummarizeCmd struct {
	*interfaces.AbstractSingleCmdByNoExecutor

	// Deps
	AllTestResults map[string][]*data.TestResults
}

// ExtractDependencies extracts all the command dependencies from state keeper.
func (cmd *SummarizeCmd) ExtractDependencies(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.PrePostFilterStateKeeper:
		err = cmd.extractDepsFromFilterStateKeepr(ctx, sk)

	default:
		return fmt.Errorf("StateKeeper '%T' is not supported by cmd type %s.", sk, cmd.GetCommandType())
	}

	if err != nil {
		return errors.Annotate(err, "error during extracting dependencies for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

// UpdateStateKeeper updates the state keeper with info from the cmd.
func (cmd *SummarizeCmd) UpdateStateKeeper(
	ctx context.Context,
	ski interfaces.StateKeeperInterface) error {

	var err error
	switch sk := ski.(type) {
	case *data.PrePostFilterStateKeeper:
		err = cmd.updateLocalTestStateKeeper(ctx, sk)
	}

	if err != nil {
		return errors.Annotate(err, "error during updating for command %s: ", cmd.GetCommandType()).Err()
	}

	return nil
}

func (cmd *SummarizeCmd) extractDepsFromFilterStateKeepr(
	ctx context.Context,
	sk *data.PrePostFilterStateKeeper) error {

	if sk.AllTestResults == nil || len(sk.AllTestResults) == 0 {
		return fmt.Errorf("Cmd %q missing dependency: AllTestResults", cmd.GetCommandType())
	}

	cmd.AllTestResults = sk.AllTestResults

	return nil
}

func (cmd *SummarizeCmd) updateLocalTestStateKeeper(
	ctx context.Context,
	sk *data.PrePostFilterStateKeeper) error {

	return nil
}

func GetResultsMapFromList(resultsList []*data.TestResults) map[string][]*data.TestResults {
	keyToResultsMap := map[string][]*data.TestResults{}

	for _, result := range resultsList {
		if _, ok := keyToResultsMap[result.Key]; !ok {
			keyToResultsMap[result.Key] = []*data.TestResults{result}
		} else {
			keyToResultsMap[result.Key] = append(keyToResultsMap[result.Key], result)
		}
	}

	return keyToResultsMap
}

// Execute executes the command.
func (cmd *SummarizeCmd) Execute(ctx context.Context) error {
	var err error
	step, ctx := build.StartStep(ctx, "Summarize")
	defer func() { step.End(err) }()

	common.WriteAnyObjectToStepLog(ctx, step, cmd.AllTestResults, "all test results")

	// sort suite keys first
	suiteKeys := make([]string, 0, len(cmd.AllTestResults))
	for k := range cmd.AllTestResults {
		suiteKeys = append(suiteKeys, k)
	}

	sort.Strings(suiteKeys)

	for _, suite := range suiteKeys {
		testResults := cmd.AllTestResults[suite]
		step, ctx := build.StartStep(ctx, suite)
		defer func() { step.End(err) }()

		// Get map from list and group error/non-error separately
		testResultsMap := GetResultsMapFromList(testResults)
		nonErrorResultKeys, nonErrorResultMap, errorResultKeys, errorResultMap := GroupErrAndNonErrResults(testResultsMap)

		errResultErr := ProcessResultsMap(ctx, errorResultKeys, errorResultMap)
		nonErrResultErr := ProcessResultsMap(ctx, nonErrorResultKeys, nonErrorResultMap)

		// Assign non nil err (if any) so that this step fails
		if errResultErr != nil {
			err = errResultErr
		} else if nonErrResultErr != nil {
			err = nonErrResultErr
		}
	}

	// we don't want the build to fail for this step
	return nil
}

func ProcessResultsMap(ctx context.Context, keys []string, resultMap map[string][]*data.TestResults) error {
	var err error
	for _, key := range keys {
		step, _ := build.StartStep(ctx, key)
		defer func() { step.End(err) }()

		resultsList := resultMap[key]
		// sort by attempt
		sort.Sort(data.ByAttempt(resultsList))
		links := []string{}

		for _, result := range resultsList {
			err = result.GetFailureErr()
			if result.TopLevelError != nil {
				DisplayError(ctx, result, step)
				continue
			}
			buildUrl := result.BuildUrl

			linkStr := "* "
			if result.Attempt > 0 {
				linkStr = fmt.Sprintf("%sretry #%d: ", linkStr, result.Attempt)
			}

			logLink := result.Results.GetLogData().GetTesthausUrl()
			if logLink != "" {
				linkStr = fmt.Sprintf("%s[log link](%s)", linkStr, logLink)
			}

			if buildUrl != "" {
				linkStr = fmt.Sprintf("%s, [task link](%s)", linkStr, buildUrl)
			}

			if linkStr != "* " {
				links = append(links, linkStr)
			}

		}
		if len(links) > 0 {
			step.SetSummaryMarkdown(strings.Join(links, "\n"))
		}
	}

	return err
}

func GroupErrAndNonErrResults(inputMap map[string][]*data.TestResults) ([]string, map[string][]*data.TestResults, []string, map[string][]*data.TestResults) {
	nonErrorResultMap := map[string][]*data.TestResults{}
	nonErrorResultKeys := []string{}
	errorResultMap := map[string][]*data.TestResults{}
	errorResultKeys := []string{}

	for _, resultList := range inputMap {
		for _, eachResult := range resultList {
			if eachResult.TopLevelError != nil {
				switch (eachResult.TopLevelError).(type) {
				case *data.EnumerationError:
					errKey := common.EnumerationErrKey
					if _, ok := errorResultMap[errKey]; !ok {
						errorResultKeys = append(errorResultKeys, errKey)
					}
					addToMap(errorResultMap, errKey, eachResult)
				case *data.BotParamsRejectedError:
					errKey := common.BotParamsRejectedErrKey
					if _, ok := errorResultMap[errKey]; !ok {
						errorResultKeys = append(errorResultKeys, errKey)
					}
					addToMap(errorResultMap, errKey, eachResult)
				default:
					errKey := common.OtherErrKey
					if _, ok := errorResultMap[errKey]; !ok {
						errorResultKeys = append(errorResultKeys, errKey)
					}
					addToMap(errorResultMap, errKey, eachResult)
				}
			} else {
				if _, ok := nonErrorResultMap[eachResult.Key]; !ok {
					nonErrorResultKeys = append(nonErrorResultKeys, eachResult.Key)
				}
				addToMap(nonErrorResultMap, eachResult.Key, eachResult)
			}
		}
	}

	sort.Strings(nonErrorResultKeys)
	sort.Strings(errorResultKeys)
	return nonErrorResultKeys, nonErrorResultMap, errorResultKeys, errorResultMap
}

func addToMap(inMap map[string][]*data.TestResults, key string, result *data.TestResults) map[string][]*data.TestResults {
	if _, ok := inMap[key]; !ok {
		inMap[key] = []*data.TestResults{}
	}
	inMap[key] = append(inMap[key], result)

	return inMap
}

func DisplayError(ctx context.Context, result *data.TestResults, step *build.Step) {
	switch e := (result.TopLevelError).(type) {
	case *data.EnumerationError:
		log := step.Log(fmt.Sprintf("suite `%s`", e.SuiteName))
		log.Write([]byte(e.Error()))
	case *data.BotParamsRejectedError:
		log := step.Log(fmt.Sprintf("bot params rejected for '%s'", e.Key))
		log.Write([]byte(fmt.Sprintf("rejected params: [\n%s\n]", strings.Join(e.RejectedDims, "\n"))))
		result.Key = common.BotParamsRejectedErrKey
	default:
		log := step.Log(fmt.Sprintf("error for '%s'", result.Key))
		log.Write([]byte(e.Error()))
	}
}

// NewSummarizeCmd returns a new SummarizeCmd
func NewSummarizeCmd() *SummarizeCmd {
	abstractCmd := interfaces.NewAbstractCmd(SummarizeCmdType)
	abstractSingleCmdByNoExecutor := &interfaces.AbstractSingleCmdByNoExecutor{AbstractCmd: abstractCmd}
	return &SummarizeCmd{AbstractSingleCmdByNoExecutor: abstractSingleCmdByNoExecutor}
}
