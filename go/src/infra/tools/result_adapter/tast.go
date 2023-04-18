// Copyright 2021 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/resultdb/pbutil"
	pb "go.chromium.org/luci/resultdb/proto/v1"
	sinkpb "go.chromium.org/luci/resultdb/sink/proto/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// The execution path for tests in Skylab envrionemnt. As of 2021Q3, all tests
	// are run inside a lxc container.
	SkylabLxcJobFolder = "/usr/local/autotest/results/lxc_job_folder"
	// The execution path for tests in CFT (F20) containers.
	CFTJobFolder = "/tmp/test/results"
	// The common name prefix for Tast test results.
	TastNamePrefix = "tast."
)

type TastResults struct {
	BaseDir string
	Cases   []TastCase
}

// Follow CrOS test platform's convention, use case to represents the single test
// executed in a Tast run. Described in
// https://pkg.go.dev/chromium.googlesource.com/chromiumos/platform/tast.git/src/chromiumos/tast/cmd/tast/internal/run/resultsjson
//
// Fields not used by Test Results are omitted.
type TastCase struct {
	Name        string           `json:"name"`
	Contacts    []string         `json:"contacts"`
	OutDir      string           `json:"outDir"`
	SkipReason  string           `json:"skipReason"`
	Errors      []TastError      `json:"errors"`
	Start       time.Time        `json:"start"`
	End         time.Time        `json:"end"`
	SearchFlags []*pb.StringPair `json:"searchFlags,omitempty"`
}

type TastError struct {
	Time   time.Time `json:"time"`
	Reason string    `json:"reason"`
	File   string    `json:"file"`
	Stack  string    `json:"stack"`
}

// ConvertFromJSON reads the provided reader into the receiver.
//
// The Cases are cleared and overwritten.
func (r *TastResults) ConvertFromJSON(reader io.Reader) error {
	r.Cases = []TastCase{}
	decoder := json.NewDecoder(reader)
	// Expected to parse JSON lines instead of a full JSON file.
	for decoder.More() {
		var t TastCase
		if err := decoder.Decode(&t); err != nil {
			return err
		}
		r.Cases = append(r.Cases, t)
	}
	return nil
}

// ToProtos converts test results in r to []*sinkpb.TestResult.
func (r *TastResults) ToProtos(ctx context.Context, testMetadataFile string, processArtifacts func(string) (map[string]string, error), testhausBaseUrl string) ([]*sinkpb.TestResult, error) {
	metadata := map[string]*api.TestCaseMetadata{}
	var err error
	if testMetadataFile != "" {
		metadata, err = parseMetadata(testMetadataFile)
		if err != nil {
			return nil, err
		}
	}

	// Convert all tast cases to TestResult.
	var ret []*sinkpb.TestResult
	for _, c := range r.Cases {
		testName := addTastPrefix(c.Name)
		status := genCaseStatus(c)
		tr := &sinkpb.TestResult{
			TestId:       testName,
			Expected:     status == pb.TestStatus_SKIP || status == pb.TestStatus_PASS,
			Status:       status,
			Tags:         []*pb.StringPair{},
			TestMetadata: &pb.TestMetadata{Name: testName},
		}

		if !c.Start.IsZero() {
			tr.StartTime = timestamppb.New(c.Start)
			if !c.End.IsZero() {
				tr.Duration = msToDuration(float64(c.End.Sub(c.Start).Milliseconds()))
			}
		}

		// Add Tags to test results.
		contacts := strings.Join(c.Contacts[:], ",")
		tr.Tags = append(tr.Tags, pbutil.StringPair("contacts", contacts))
		tr.Tags = append(tr.Tags, c.SearchFlags...)

		testMetadata, ok := metadata[testName]
		if ok {
			tr.Tags = append(tr.Tags, metadataToTags(testMetadata)...)
			tr.TestMetadata.BugComponent, err = parseBugComponentMetadata(testMetadata)
			if err != nil {
				logging.Errorf(
					ctx,
					"could not parse bug component metadata from: %v due to: %v",
					testMetadata,
					err)
			}
		}

		if status == pb.TestStatus_SKIP {
			tr.SummaryHtml = "<text-artifact artifact-id=\"Skip Reason\" />"
			tr.Artifacts = map[string]*sinkpb.Artifact{
				"Skip Reason": {
					Body:        &sinkpb.Artifact_Contents{Contents: []byte(c.SkipReason)},
					ContentType: "text/plain",
				}}
			ret = append(ret, tr)
			continue
		}

		d := c.OutDir
		tr.Artifacts = map[string]*sinkpb.Artifact{}
		// For Skylab tests, the OutDir recorded by tast is different from the
		// result folder we can access on Drone server.
		if strings.HasPrefix(d, SkylabLxcJobFolder) {
			d = strings.Replace(d, SkylabLxcJobFolder, r.BaseDir, 1)
		} else if strings.HasPrefix(d, CFTJobFolder) {
			d = strings.Replace(d, CFTJobFolder, r.BaseDir, 1)
		}
		normPathToFullPath, err := processArtifacts(d)
		if err != nil {
			return nil, err
		}
		for f, p := range normPathToFullPath {
			tr.Artifacts[f] = &sinkpb.Artifact{
				Body: &sinkpb.Artifact_FilePath{FilePath: p},
			}
		}

		if testhausBaseUrl != "" {
			tr.Artifacts["testhaus_logs"] = &sinkpb.Artifact{
				Body: &sinkpb.Artifact_Contents{
					Contents: []byte(fmt.Sprintf("%s/cros-test/artifact/tast/tests/%s", strings.TrimSuffix(testhausBaseUrl, "/"), c.Name)),
				},
				ContentType: "text/x-uri",
			}
		}

		if len(c.Errors) > 0 {
			tr.FailureReason = &pb.FailureReason{
				PrimaryErrorMessage: truncateString(c.Errors[0].Reason, maxPrimaryErrorBytes),
			}
			errLog := ""
			for _, err := range c.Errors {
				errLog += err.Stack
				errLog += "\n"
			}
			tr.Artifacts["Test Log"] = &sinkpb.Artifact{
				Body:        &sinkpb.Artifact_Contents{Contents: []byte(errLog)},
				ContentType: "text/plain",
			}
			tr.SummaryHtml = "<text-artifact artifact-id=\"Test Log\" />"
		}
		ret = append(ret, tr)
	}
	return ret, nil
}

func addTastPrefix(testName string) string {
	if strings.HasPrefix(testName, TastNamePrefix) {
		return testName
	}
	return TastNamePrefix + testName
}

func genCaseStatus(c TastCase) pb.TestStatus {
	if c.SkipReason != "" {
		return pb.TestStatus_SKIP
	}
	if len(c.Errors) > 0 {
		return pb.TestStatus_FAIL
	}
	return pb.TestStatus_PASS
}
