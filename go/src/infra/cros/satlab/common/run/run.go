// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/golang/protobuf/jsonpb"
	"github.com/googleapis/gax-go/v2"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/platform/dev-util/src/chromiumos/ctp/builder"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/api/option"
	moblabpb "google.golang.org/genproto/googleapis/chromeos/moblab/v1beta1"

	"infra/cros/satlab/common/google.golang.org/google/chromeos/moblab"
	"infra/cros/satlab/common/satlabcommands"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/executor"
)

// Run holds the arguments that are needed for the run command.
type Run struct {
	Image         string
	Model         string
	Board         string
	Milestone     string
	Build         string
	Pool          string
	Suite         string
	Tests         []string
	Testplan      string
	TestplanLocal string
	Harness       string
	TestArgs      string
	SatlabId      string
	CFT           bool
	Local         bool
	MaxTimeout    bool
	// Any configs related to results upload for this test run.
	AddedDims map[string]string
	Tags      map[string]string
}

// TriggerRun triggers the Run with the given information
// (it could be either single test or a suite or a test_plan in the GCS bucket or test_plan saved locally)
func (c *Run) TriggerRun(ctx context.Context) (string, error) {
	bbClient, err := c.createCTPBuilder(ctx)
	if err != nil {
		return "", err
	}
	// Create default client
	err = bbClient.AddDefaultBBClient(ctx)
	if err != nil {
		return "", err
	}

	moblabClient, err := moblab.NewBuildClient(ctx, option.WithCredentialsFile(site.GetServiceAccountPath()))
	if err != nil {
		return "", errors.Annotate(err, "satlab new moblab api build client").Err()
	}

	link, err := c.triggerRunWithClients(ctx, moblabClient, bbClient, site.GetGCSImageBucket())
	if err != nil {
		return "", errors.Annotate(err, "triggerRunWithClients").Err()
	}
	return link, nil
}

func (c *Run) createCTPBuilder(ctx context.Context) (*builder.CTPBuilder, error) {
	// Create TestPlan for suite or test
	tp, err := c.createTestPlan()
	if err != nil {
		return nil, err
	}
	// Set tags to pass to ctp and test runner builds
	c.Tags = c.setTags()

	dims := c.AddedDims
	// Will be nil if not provided by user.
	if dims == nil {
		dims = make(map[string]string)
	}

	// Get drone target based on user input, defaulting to the current box.
	droneDim, err := c.getDroneTarget(ctx)
	if err != nil {
		return nil, err
	}
	if droneDim != "" {
		dims["drone"] = droneDim
	}

	builderId := &buildbucketpb.BuilderID{
		Project: site.GetLUCIProject(),
		Bucket:  site.GetCTPBucket(),
		Builder: site.GetCTPBuilder(),
	}

	if c.Image == "" {
		c.Image = fmt.Sprintf("%s-release/R%s-%s", c.Board, c.Milestone, c.Build)
	}
	bbClient := &builder.CTPBuilder{
		Image:               c.Image,
		Board:               c.Board,
		Model:               c.Model,
		Pool:                c.Pool,
		CFT:                 c.CFT,
		TestPlan:            tp,
		BuilderID:           builderId,
		Dimensions:          dims,
		ImageBucket:         site.GetGCSImageBucket(),
		AuthOptions:         &site.DefaultAuthOptions,
		TestRunnerBuildTags: c.Tags,
		TimeoutMins:         c.setTimeout(),
		// TRV2:        true,
	}

	return bbClient, nil
}

func (c *Run) triggerRunWithClients(ctx context.Context, moblabClient MoblabClient, bbClient BuildbucketClient, gcsBucket string) (string, error) {

	// There is no explicit check on whether staging of the image is successful or not
	// There are 2 reasons for this:
	// 1. "Custom chromeOS builds" are expected to be already in the partner bucket. There is no
	// check on whether that already exists in the bucket. (In an ideal world, there would be
	// one, but right now there is none. This is much harder because there is no list of
	// compulsory artifacts that should exist in the folder)
	// 2. Latency: Waiting for the copying to take place is not a good user experience and
	// is not necessary anyway in this case. Although copying is fairly quick, it is left to
	// be handled by server in the background
	err := StageImageToBucket(ctx, moblabClient, c.Board, c.Model, c.Build)
	if err != nil {
		return "", errors.Annotate(err, "satlab stage image to bucket").Err()
	}

	link, err := ScheduleBuild(ctx, bbClient)
	if err != nil {
		return "", errors.Annotate(err, "satlab schedule build").Err()
	}
	return link, nil
}

// SetTags passes testplan name as a tag for associated tests and suites
func (c *Run) setTags() map[string]string {
	tags := make(map[string]string)

	if c.Testplan != "" {
		tags["test-plan-id"] = strings.TrimSuffix(c.Testplan, ".json")
	} else if c.TestplanLocal != "" {
		tags["test-plan-id"] = strings.TrimSuffix(c.TestplanLocal, ".json")
	}
	return tags
}

// Determine CTP timeout based on user input
func (c *Run) setTimeout() int {
	if c.MaxTimeout {
		return 2370
	}
	return 360
}

func (c *Run) createTestPlan() (*test_platform.Request_TestPlan, error) {
	var tp *test_platform.Request_TestPlan
	var err error

	if c.Suite != "" {
		tp = builder.TestPlanForSuites([]string{c.Suite})
	} else if c.Tests != nil {
		tp = builder.TestPlanForTests(c.TestArgs, c.Harness, c.Tests)
	} else if c.Testplan != "" {
		fmt.Printf("Fetching testplan...\n")
		var w bytes.Buffer
		path, err := downloadTestPlan(&w, site.GetGCSImageBucket(), c.Testplan)
		if err != nil {
			return nil, err
		}
		tp, err = readTestPlan(path)
		if err != nil {
			return nil, err
		}
	} else if c.TestplanLocal != "" {
		tp, err = readTestPlan(c.TestplanLocal)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Running local testplan...\n")
	}

	return tp, nil
}

// StageImageToBucket stages the specified Chrome OS image to the user GCS bucket
func StageImageToBucket(ctx context.Context, moblabClient MoblabClient, board string, model string, buildVersion string) error {
	bucket := site.GetGCSImageBucket()
	if bucket == "" {
		return errors.New("GCS_BUCKET not found")
	}

	buildTarget := fmt.Sprintf("buildTargets/%s/models/%s", board, model)
	artifactName := fmt.Sprintf("%s/builds/%s/artifacts/%s", buildTarget, buildVersion, bucket)
	stageReq := &moblabpb.StageBuildRequest{
		Name: artifactName,
	}

	_, err := moblabClient.StageBuild(ctx, stageReq)
	if err != nil {
		return err
	}
	var stageStatus *moblabpb.CheckBuildStageStatusResponse
	count := 10
	for {
		count--
		req := &moblabpb.CheckBuildStageStatusRequest{
			Name: artifactName,
		}
		stageStatus, err = moblabClient.CheckBuildStageStatus(ctx, req)
		if err != nil {
			return err
		}
		if stageStatus.IsBuildStaged {
			break
		}
		if count == 0 {
			return fmt.Errorf("stage not completed within 10 retries")
		}
	}
	destPath := stageStatus.StagedBuildArtifact.Path

	fmt.Printf("Artifacts staged to %s\n", path.Join(bucket, destPath))
	return nil
}

// / ScheduleBuild register a build. If it successes, it returns a link of build. Otherwise,
// / return an error.
func ScheduleBuild(ctx context.Context, bbClient BuildbucketClient) (string, error) {
	ctpBuild, err := bbClient.ScheduleCTPBuild(ctx)
	if err != nil {
		return "", err
	}
	link := fmt.Sprintf("https://ci.chromium.org/ui/b/%s", strconv.Itoa(int(ctpBuild.Id)))
	return link, nil
}

// Set drone target to user-provided satlab or local satlab if one isn't provided
func (c *Run) getDroneTarget(ctx context.Context) (string, error) {
	var satlabTarget string
	if c.SatlabId != "" {
		satlabTarget = fmt.Sprintf(c.SatlabId)
	} else if c.Local { // get id of local satlab if one is not provided
		localSatlab, err := satlabcommands.GetDockerHostBoxIdentifier(ctx, &executor.ExecCommander{})
		if err != nil {
			return "", errors.Annotate(err, "satlab get docker host box identifier").Err()
		}
		satlabTarget = fmt.Sprintf("satlab-%s", localSatlab)
	}
	return satlabTarget, nil
}

// BuildbucketClient interface provides subset of Buildbucket methods relevant to Satlab CLI
type BuildbucketClient interface {
	ScheduleCTPBuild(ctx context.Context) (*buildbucketpb.Build, error)
}

// MoblabClient interface provides subset of Moblab API methods relevant to Satlab CLI
type MoblabClient interface {
	StageBuild(ctx context.Context, req *moblabpb.StageBuildRequest, opts ...gax.CallOption) (*moblab.StageBuildOperation, error)
	CheckBuildStageStatus(ctx context.Context, req *moblabpb.CheckBuildStageStatusRequest, opts ...gax.CallOption) (*moblabpb.CheckBuildStageStatusResponse, error)
}

// Downloads specified testplan from bucket to remote access container
func downloadTestPlan(w io.Writer, bucket, testplan string) (string, error) {
	object := "testplans/" + testplan
	destFileName := "/config/" + testplan
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(site.GetServiceAccountPath()))
	if err != nil {
		return "", fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("%q Error: %w", object, err)
	}
	defer rc.Close()

	f, err := os.Create(destFileName)
	if err != nil {
		return "", fmt.Errorf("os.Create: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return "", fmt.Errorf("io.Copy: %w", err)
	}

	fmt.Fprintf(w, "Blob %v downloaded to local file %v\n", object, destFileName)

	return destFileName, nil
}

// JSONPBUnmarshaler unmarshals JSON into proto messages.
var JSONPBUnmarshaler = jsonpb.Unmarshaler{AllowUnknownFields: true}

func readTestPlan(path string) (*test_platform.Request_TestPlan, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error reading test plan: %v", err)
	}
	defer file.Close()

	testPlan := &test_platform.Request_TestPlan{}
	if err := JSONPBUnmarshaler.Unmarshal(file, testPlan); err != nil {
		return nil, fmt.Errorf("error reading test plan: %v", err)
	}
	return testPlan, nil
}
