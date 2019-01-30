// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package step

import (
	"fmt"
	"sort"
	"strings"

	"go.chromium.org/luci/common/logging"

	"golang.org/x/net/context"

	"infra/appengine/sheriff-o-matic/som/client"
	te "infra/appengine/sheriff-o-matic/som/testexpectations"
	"infra/appengine/test-results/model"
	"infra/monitoring/messages"
)

// Max number of failed tests to put in resulting json. Needed because of datastore size limits.
var maxFailedTests = 40

// Text to put in test names indicating results were snipped.
const tooManyFailuresText = "...... too many results, data snipped...."

// TestFailure is a failure of a single test suite. May include multiple test cases.
// Can also include information about failure causes, including findit information.
type TestFailure struct {
	// Could be more detailed about test failures. For instance, we could
	// indicate expected vs. actual result.
	//FIXME: Merge TestNames and Tests.
	TestNames []string `json:"test_names"`
	//FIXME: Rename to TestSuite (needs to be synchronized with SOM)
	StepName string           `json:"step"`
	Tests    []TestWithResult `json:"tests"`
	// For test-results in SoM
	AlertTestResults []messages.AlertTestResults `json:"alert_test_results"`
}

// Signature implements the ReasonRaw interface
func (t *TestFailure) Signature() string {
	if len(t.TestNames) == 0 {
		return t.StepName
	}
	return strings.Join(t.TestNames, ",")
}

// testTrunc returns the test name, potentially truncated.
func (t *TestFailure) testTrunc() string {
	if len(t.TestNames) > 1 {
		return fmt.Sprintf("%s and %d other(s)", t.TestNames[0], len(t.TestNames)-1)
	}

	split := strings.Split(t.TestNames[0], "/")
	// If the test name has a URL in it with https:, don't split it up and truncate.
	// If we did try that, the test name would be basically nonsense. Just show the full
	// thing, even if it's a bit large.
	if len(split) > 2 && !(contains(split, "https:") || contains(split, "http:")) {
		split = []string{split[0], "...", split[len(split)-1]}
	}
	return strings.Join(split, "/")
}

// Kind implements the ReasonRaw interface
func (t *TestFailure) Kind() string {
	return "test"
}

// Severity implements the ReasonRaw interface
func (t *TestFailure) Severity() messages.Severity {
	return messages.NoSeverity
}

// Title implements the ReasonRaw interface
func (t *TestFailure) Title(bses []*messages.BuildStep) string {
	f := bses[0]
	name := GetTestSuite(f)
	if len(t.TestNames) > 0 {
		name = t.testTrunc() + " in " + name
	}

	if len(bses) == 1 {
		return fmt.Sprintf("%s failing on %s/%s", name, f.Master.Name(), f.Build.BuilderName)
	}

	return fmt.Sprintf("%s failing on multiple builders", name)
}

// ArtifactLink is a link to a test artifact left by perf tests.
type ArtifactLink struct {
	// Name is the name of the artifact.
	Name string `json:"name"`
	// Location is the location of the artifact.
	Location string `json:"location"`
}

// TestWithResult stores the information provided by Findit for a specific test,
// for example if the test is flaky or is there a culprit for the test failure.
// Also contains test-specific details like expectations and any artifacts
// produced by the test run.
type TestWithResult struct {
	TestName     string                     `json:"test_name"`
	IsFlaky      bool                       `json:"is_flaky"`
	SuspectedCLs []messages.SuspectCL       `json:"suspected_cls"`
	Expectations []*te.ExpectationStatement `json:"expectations"`
	Artifacts    []ArtifactLink             `json:"artifacts"`
}

type testFailureAnalyzer struct {
	findit client.FindIt
	trc    client.TestResults
}

// testFailureAnalyzer analyzes steps to see if there is any data in the tests
// server which corresponds to the failure.
func (tfa *testFailureAnalyzer) Analyze(ctx context.Context, fs []*messages.BuildStep, tree string) ([]messages.ReasonRaw, []error) {
	results := make([]messages.ReasonRaw, len(fs))
	builderConfigs, err := te.LoadBuilderConfigs(ctx)
	if err != nil {
		return nil, []error{err}
	}

	for i, f := range fs {
		rslt, err := testAnalyzeFailure(ctx, f, tfa.findit, tfa.trc)
		if err != nil {
			return nil, []error{err}
		}
		if rslt == nil {
			continue
		}
		failure, ok := rslt.(*TestFailure)
		if !ok {
			logging.Errorf(ctx, "couldn't cast to *TestFailure: %+v", rslt)
			continue
		}
		for t, r := range failure.Tests {
			if tree == "chromium.perf" {
				artifacts, err := getArtifactsForTest(ctx, f, r.TestName, tfa.trc)
				if err != nil {
					logging.Errorf(ctx, "couldn't get test artifacts for %+v: %v", r.TestName, err)
				} else {
					failure.Tests[t].Artifacts = artifacts
					logging.Infof(ctx, "added %d artifacts to %q", len(failure.Tests[t].Artifacts), r.TestName)
				}
			}

			config, ok := builderConfigs[f.Build.BuilderName]
			if !ok {
				logging.Warningf(ctx, "no config (out of %d) for %s", len(builderConfigs), f.Build.BuilderName)
				continue
			}
			exps, err := getExpectationsForTest(ctx, r.TestName, config)
			if err != nil {
				logging.Errorf(ctx, "couldn't get test expectations for %+v: %v", r.TestName, err)
			} else {
				failure.Tests[t].Expectations = exps
			}
		}
		results[i] = failure
	}

	return results, nil
}

func getArtifactsForTest(ctx context.Context, f *messages.BuildStep, name string, trc client.TestResults) ([]ArtifactLink, error) {
	ret := []ArtifactLink{}
	suiteName := GetTestSuite(f)
	testResults, err := trc.TestResults(ctx, f.Master, f.Build.BuilderName, suiteName, f.Build.Number)
	if err != nil {
		return nil, err
	}
	if testResults == nil {
		return nil, fmt.Errorf("didn't get any test results for test name %q", name)
	}
	delim := "/"
	if testResults.PathDelim != nil {
		delim = *testResults.PathDelim
	}

	for testName, res := range testResults.Tests.Flatten(delim) {
		if testName == name {
			for artName, locations := range res.Artifacts {
				for _, loc := range locations {
					ret = append(ret, ArtifactLink{
						Name:     artName,
						Location: loc,
					})
				}
			}
		}
	}

	return ret, nil
}

// tests is a slice of tests with Findit results.
type tests []TestWithResult

func (slice tests) Len() int {
	return len(slice)
}

func (slice tests) Less(i, j int) bool {
	return (len(slice[i].SuspectedCLs) > 0 && len(slice[j].SuspectedCLs) == 0) || (slice[i].IsFlaky && !slice[j].IsFlaky)
}

func (slice tests) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func testAnalyzeFailure(ctx context.Context, f *messages.BuildStep, findit client.FindIt, trc client.TestResults) (messages.ReasonRaw, error) {
	suiteName, failedTests, err := getTestNames(ctx, f, trc)
	if err != nil {
		return nil, err
	}

	testsWithFinditResults, err := getFinditResultsForTests(ctx, f, failedTests, findit)
	if err != nil {
		logging.Errorf(ctx, "Error calling Findit, but continuing: %v", err)
	}

	if failedTests != nil {
		sortedNames := failedTests
		sort.Strings(sortedNames)
		sortedTests := tests(testsWithFinditResults)
		sort.Sort(sortedTests)
		return &TestFailure{
			TestNames: sortedNames,
			StepName:  suiteName,
			Tests:     testsWithFinditResults,
		}, nil
	}

	return nil, nil
}

// GetTestSuite returns the name of the test suite executed in a step. Currently, it has
// a bunch of custom logic to parse through all the suffixes added by various recipe code.
// Eventually, it should just read something structured from the step.
// https://bugs.chromium.org/p/chromium/issues/detail?id=674708
func GetTestSuite(bs *messages.BuildStep) string {
	testSuite := bs.Step.Name
	s := strings.Split(bs.Step.Name, " ")

	if bs.Master.Name() == "chromium.perf" {
		found := false
		// If a step has a swarming.summary log, then we assume it's a test
		for _, b := range bs.Step.Logs {
			if len(b) > 1 && b[0] == "swarming.summary" {
				found = true
				break
			}
		}

		if !found {
			return ""
		}
	} else if !(strings.HasSuffix(s[0], "tests") || strings.HasSuffix(s[0], "test_apk")) {
		// Some test steps have names like "webkit_tests iOS(dbug)" so we look at the first
		// term before the space, if there is one.
		return testSuite
	}

	// Recipes add a suffix to steps of the OS that it's run on, when the test
	// is swarmed. The step name is formatted like this: "<task title> on <OS>".
	// Added in this code:
	// https://chromium.googlesource.com/chromium/tools/build/+/9ef66559727c320b3263d7e82fb3fcd1b6a3bd55/scripts/slave/recipe_modules/swarming/api.py#846
	if len(s) > 2 && s[1] == "on" {
		testSuite = s[0]
	}

	return testSuite
}

func getTestNames(ctx context.Context, f *messages.BuildStep, trc client.TestResults) (string, []string, error) {
	name := GetTestSuite(f)
	if name == "" {
		return name, nil, nil
	}

	testResults, err := trc.TestResults(ctx, f.Master, f.Build.BuilderName, name, f.Build.Number)

	if testResults == nil || len(testResults.Tests) == 0 || err != nil {
		if err != nil {
			// Still want to keep on serving some data, even if test results is down. We can also do something
			// in this analyzer, even if we have no test results.
			logging.Warningf(ctx, "got error fetching test results (ignoring): %s", err)
		}

		if name != f.Step.Name {
			logging.Infof(ctx, "name != f.Step.Name: %q vs %q", name, f.Step.Name)
			// Signal that we still found something useful, even if we
			// don't have test results.
			return name, []string{}, nil
		}

		return name, nil, nil
	}

	failedTests := unexpectedFailures(testResults)

	if len(failedTests) > maxFailedTests {
		sort.Strings(failedTests)
		logging.Errorf(ctx, "Too many failed tests (%d) to put in the resulting json.", len(failedTests))
		failedTests = append(failedTests[:maxFailedTests], tooManyFailuresText)
	}

	return name, failedTests, nil
}

func unexpectedFailures(testResults *model.FullResult) []string {
	failedTests := []string{}
	delim := "/"
	if testResults.PathDelim != nil {
		delim = *testResults.PathDelim
	}

	for testName, res := range testResults.Tests.Flatten(delim) {
		if res.Unexpected != nil && *res.Unexpected {
			hasPass := false
			for _, act := range res.Actual {
				if act == "PASS" {
					hasPass = true
					break
				}
			}
			if !hasPass {
				failedTests = append(failedTests, testName)
			}
		}
	}
	return failedTests
}

// Read Findit results and get suspected cls or check if flaky for each test.
func getFinditResultsForTests(ctx context.Context, f *messages.BuildStep, failedTests []string, findit client.FindIt) ([]TestWithResult, error) {
	TestsWithFinditResults := []TestWithResult{}

	if failedTests == nil || len(failedTests) == 0 {
		return nil, nil
	}

	name := GetTestSuite(f)
	if name == "" {
		return nil, nil
	}

	finditResults, err := findit.Findit(ctx, f.Master, f.Build.BuilderName, f.Build.Number, []string{name})
	if err != nil {
		logging.Warningf(ctx, "ignoring findit error: %s", err)
	}
	finditResultsMap := map[string]*messages.FinditResult{}
	for _, result := range finditResults {
		finditResultsMap[result.TestName] = result
	}
	for _, test := range failedTests {
		testResult := TestWithResult{
			TestName:     test,
			IsFlaky:      false,
			SuspectedCLs: nil,
		}
		result, ok := finditResultsMap[test]
		if ok {
			testResult = TestWithResult{
				TestName:     test,
				IsFlaky:      result.IsFlakyTest,
				SuspectedCLs: result.SuspectedCLs,
			}
		}
		TestsWithFinditResults = append(TestsWithFinditResults, testResult)
	}
	return TestsWithFinditResults, nil
}

func getExpectationsForTest(ctx context.Context, testName string, config *te.BuilderConfig) ([]*te.ExpectationStatement, error) {
	fs, err := te.LoadAll(ctx)
	if err != nil {
		return nil, err
	}

	return fs.ForTest(testName, config), nil
}

func contains(haystack []string, needle string) bool {
	for _, itm := range haystack {
		if itm == needle {
			return true
		}
	}

	return false
}
