// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/resultdb/pbutil"
	pb "go.chromium.org/luci/resultdb/proto/v1"
	sinkpb "go.chromium.org/luci/resultdb/sink/proto/v1"
)

var (
	// Prefixes that may be present in the test name and must be stripped before forming the base id.
	prefixes = []string{"MANUAL_", "PRE_"}

	// Java base ids aren't actually GTest but use the same launcher output format.
	// TODO(chanli@): remove java test related logic when crbug.com/1104245 and crbug.com/1104238 are done.
	javaIDRe = regexp.MustCompile(`^[\w.]+#`)

	// Test base ids look like FooTest.DoesBar: "FooTest" is the suite and "DoesBar" the test name.
	baseIDRE = regexp.MustCompile(`^(\w+)\.(\w+)$`)

	// Type parametrized test examples:
	// - MyInstantiation/FooTest/1.DoesBar
	// - FooTest/1.DoesBar
	// - FooType/MyType.DoesBar
	//
	// In the above examples, "FooTest" is the suite, "DoesBar" the test name, "MyInstantiation" the
	// optional instantiation, "1" the index of the type on which the test has been instantiated, if
	// no string representation for the type has been provided, and "MyType" is the user-provided
	// string representation of the type on which the test has been instantiated.
	typeParamRE = regexp.MustCompile(`^((\w+)/)?(\w+)/(\w+)\.(\w+)$`)

	// Value parametrized tests examples:
	// - MyInstantiation/FooTest.DoesBar/1
	// - FooTest.DoesBar/1
	// - FooTest.DoesBar/TestValue
	//
	// In the above examples, "FooTest" is the suite, "DoesBar" the test name, "MyInstantiation" the
	// optional instantiation, "1" the index of the value on which the test has been instantiated, if
	// no string representation for the value has been provided, and "TestValue" is the user-provided
	// string representation of the value on which the test has been instantiated.
	valueParamRE = regexp.MustCompile(`^((\w+)/)?(\w+)\.(\w+)/(\w+)$`)

	// TODO(chanli@): Remove this after crbug.com/1045846 is fixed.
	// This is a synthetic test created by test launcher, not a real test.
	syntheticTestRE = regexp.MustCompile(`^GoogleTestVerification.Uninstantiated(?:Type)?ParamaterizedTestSuite<\w+>$`)

	syntheticTestTag = errors.BoolTag{
		Key: errors.NewTagKey("synthetic test"),
	}
)

// GTestResults represents the structure as described to be generated in
// https://cs.chromium.org/chromium/src/base/test/launcher/test_results_tracker.h?l=83&rcl=96020cfd447cb285acfa1a96c37a67ed22fa2499
// (base::TestResultsTracker::SaveSummaryAsJSON)
//
// Fields not used by Test Results are omitted.
type GTestResults struct {
	AllTests      []string `json:"all_tests"`
	DisabledTests []string `json:"disabled_tests"`
	GlobalTags    []string `json:"global_tags"`

	// PerIterationData is a vector of run iterations, each mapping test names to a list of test data.
	PerIterationData []map[string][]*GTestRunResult `json:"per_iteration_data"`

	// TestLocations maps test names to their location in code.
	TestLocations map[string]*Location `json:"test_locations"`
	buf           *strings.Builder
}

// GTestRunResult represents the per_iteration_data as described in
// https://cs.chromium.org/chromium/src/base/test/launcher/test_results_tracker.h?l=83&rcl=96020cfd447cb285acfa1a96c37a67ed22fa2499
// (base::TestResultsTracker::SaveSummaryAsJSON)
//
// Fields not used by Test Results are omitted.
type GTestRunResult struct {
	Status        string  `json:"status"`
	ElapsedTimeMs float64 `json:"elapsed_time_ms"`

	LosslessSnippet     bool   `json:"losless_snippet"`
	OutputSnippetBase64 string `json:"output_snippet_base64"`

	// Links are not generated by test_launcher, but harnesses built on top may add them to the json.
	Links map[string]json.RawMessage `json:"links"`
}

// Location describes a code location.
type Location struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// Link represents a link content.
type Link struct {
	Content string `json:"content"`
}

// ConvertFromJSON reads the provided reader into the receiver.
//
// The receiver is cleared and its fields overwritten.
func (r *GTestResults) ConvertFromJSON(reader io.Reader) error {
	*r = GTestResults{
		buf: &strings.Builder{},
	}
	if err := json.NewDecoder(reader).Decode(r); err != nil {
		return err
	}

	return nil
}

// ToProtos converts test results in r to []*sinkpb.TestResult.
func (r *GTestResults) ToProtos(ctx context.Context) ([]*sinkpb.TestResult, error) {
	var ret []*sinkpb.TestResult
	var testNames []string

	globalTags := make([]*pb.StringPair, len(r.GlobalTags)+1)
	for i, tag := range r.GlobalTags {
		globalTags[i] = pbutil.StringPair("gtest_global_tag", tag)
	}
	globalTags[len(r.GlobalTags)] = pbutil.StringPair(originalFormatTagKey, formatGTest)

	for _, name := range r.DisabledTests {
		testID, err := extractGTestParameters(name)
		switch {
		case syntheticTestTag.In(err):
			continue
		case err != nil:
			return nil, errors.Annotate(err,
				"failed to extract test id and parameters from %q", name).Err()
		}
		tr := &sinkpb.TestResult{
			TestId:   testID,
			Expected: true,
			Status:   pb.TestStatus_SKIP,
			Tags: pbutil.StringPairs(
				// Store the original Gtest test name.
				"test_name", name,
				"disabled_test", "true",
			),
			TestMetadata: &pb.TestMetadata{Name: name},
		}
		tr.Tags = append(tr.Tags, globalTags...)

		ret = append(ret, tr)
	}

	for _, data := range r.PerIterationData {
		// Sort the test name to make the output deterministic.
		testNames = testNames[:0]
		for name := range data {
			testNames = append(testNames, name)
		}
		sort.Strings(testNames)

		for _, name := range testNames {
			testID, err := extractGTestParameters(name)
			switch {
			case syntheticTestTag.In(err):
				continue
			case err != nil:
				return nil, errors.Annotate(err,
					"failed to extract test id and parameters from %q", name).Err()
			}

			for i, result := range data[name] {
				// Store the processed test result into the correct part of the overall map.
				rpb, err := r.convertTestResult(ctx, testID, name, result)
				if err != nil {
					return nil, errors.Annotate(err,
						"iteration %d of test %s failed to convert run result", i, name).Err()
				}
				rpb.Tags = append(rpb.Tags, globalTags...)

				ret = append(ret, rpb)
			}
		}
	}

	return ret, nil
}

func fromGTestStatus(s string) (status pb.TestStatus, expected bool, err error) {
	switch s {
	case "SUCCESS":
		return pb.TestStatus_PASS, true, nil
	case "FAILURE":
		return pb.TestStatus_FAIL, false, nil
	case "FAILURE_ON_EXIT":
		return pb.TestStatus_FAIL, false, nil
	case "TIMEOUT":
		return pb.TestStatus_ABORT, false, nil
	case "CRASH":
		return pb.TestStatus_CRASH, false, nil
	case "SKIPPED":
		return pb.TestStatus_SKIP, true, nil
	case "EXCESSIVE_OUTPUT":
		return pb.TestStatus_FAIL, false, nil
	case "NOTRUN":
		return pb.TestStatus_SKIP, false, nil
	case "UNKNOWN":
		return pb.TestStatus_ABORT, false, nil
	default:
		// This would only happen if the set of possible GTest result statuses change and resultsdb has
		// not been updated to match.
		return pb.TestStatus_STATUS_UNSPECIFIED, false, errors.Reason("unknown GTest status %q", s).Err()
	}
}

// extractGTestParameters extracts parameters from a test id as a mapping with "param/" keys.
func extractGTestParameters(testID string) (baseID string, err error) {
	var suite, name, instantiation, id string

	// If this is a JUnit tests, don't try to extract parameters.
	if match := javaIDRe.FindStringSubmatch(testID); match != nil {
		baseID = testID
		return
	}

	// Tests can be only one of type- or value-parametrized, if parametrized at all.
	if match := typeParamRE.FindStringSubmatch(testID); match != nil {
		// Extract type parameter.
		suite = match[3]
		name = match[5]
		instantiation = match[2]
		id = match[4]
	} else if match := valueParamRE.FindStringSubmatch(testID); match != nil {
		// Extract value parameter.
		suite = match[3]
		name = match[4]
		instantiation = match[2]
		id = match[5]
	} else if match := baseIDRE.FindStringSubmatch(testID); match != nil {
		// Otherwise our test id should not be parametrized, so extract the suite
		// and name.
		suite = match[1]
		name = match[2]
	} else if syntheticTestRE.MatchString(testID) {
		// A synthetic test, skip.
		err = errors.Reason("not a real test").Tag(syntheticTestTag).Err()
	} else {
		// Otherwise test id format is unrecognized.
		err = errors.Reason("test id of unknown format").Err()
		return
	}

	// Strip prefixes from test name if necessary.
	name = stripRepeatedPrefixes(name, prefixes...)

	switch {
	case id == "":
		baseID = fmt.Sprintf("%s.%s", suite, name)
	case instantiation == "":
		baseID = fmt.Sprintf("%s.%s/%s", suite, name, id)
	default:
		baseID = fmt.Sprintf("%s.%s/%s.%s", suite, name, instantiation, id)
	}

	return
}

func (r *GTestResults) convertTestResult(ctx context.Context, testID, name string, result *GTestRunResult) (*sinkpb.TestResult, error) {
	status, expected, err := fromGTestStatus(result.Status)
	if err != nil {
		return nil, err
	}

	tr := &sinkpb.TestResult{
		TestId:   testID,
		Expected: expected,
		Status:   status,
		Tags: pbutil.StringPairs(
			// Store the original Gtest test name.
			"test_name", name,
			// Store the original GTest status.
			"gtest_status", result.Status,
			// Store the correct output snippet.
			"lossless_snippet", strconv.FormatBool(result.LosslessSnippet),
		),
		TestMetadata: &pb.TestMetadata{Name: name},
	}

	// Do not set duration if it is unknown.
	if result.ElapsedTimeMs != 0 {
		tr.Duration = msToDuration(result.ElapsedTimeMs)
	}

	summaryData := map[string]interface{}{}

	// snippet
	if result.OutputSnippetBase64 != "" {
		outputBytes, err := base64.StdEncoding.DecodeString(result.OutputSnippetBase64)
		if err != nil {
			// Log the error, but we shouldn't fail to convert a file just because we can't
			// convert a summary.
			logging.Warningf(ctx, "Failed to convert OutputSnippetBase64 %q", result.OutputSnippetBase64)
		} else {
			tr.Artifacts = map[string]*sinkpb.Artifact{"snippet": {
				Body:        &sinkpb.Artifact_Contents{Contents: outputBytes},
				ContentType: "text/plain",
			}}
			summaryData["snippet"] = `<p><text-artifact artifact-id="snippet" /></p>`
		}
	}

	// links
	if len(result.Links) > 0 {
		links := make(map[string]string, len(result.Links))
		l := new(Link)
		s := new(string)

		for lName, link := range result.Links {
			switch {
			case json.Unmarshal(link, l) == nil:
				links[lName] = l.Content
			case json.Unmarshal(link, s) == nil:
				links[lName] = *s
			default:
				return nil, errors.Reason("unsupported data format for a link: %q", string(link)).Err()
			}
		}
		summaryData["links"] = links
	}

	// Write the summary html
	if len(summaryData) > 0 {
		r.buf.Reset()
		err = summaryTmpl.ExecuteTemplate(r.buf, "gtest", summaryData)
		if err != nil {
			return nil, err
		}
		tr.SummaryHtml = r.buf.String()
	}

	// Store the test code location.
	if loc, ok := r.TestLocations[name]; ok {
		file := normalizePath(loc.File)
		// For some reason, many file paths start with "../../", followed by
		// the correct path. Strip the prefix.
		file = stripRepeatedPrefixes(file, "../")
		file = ensureLeadingDoubleSlash(file)
		tr.TestMetadata.Location = &pb.TestLocation{
			Repo:     chromiumSrcRepo,
			FileName: file,
			Line:     int32(loc.Line),
		}
	}

	return tr, nil
}

// stripRepeatedPrefixes strips prefixes.
func stripRepeatedPrefixes(str string, prefixes ...string) string {
	for {
		stripped := str
		for _, prefix := range prefixes {
			stripped = strings.TrimPrefix(stripped, prefix)
		}
		if stripped == str {
			return str
		}
		str = stripped
	}
}
