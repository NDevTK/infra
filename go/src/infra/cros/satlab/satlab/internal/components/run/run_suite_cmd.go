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
	"github.com/maruel/subcommands"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/platform/dev-util/src/chromiumos/ctp/builder"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/api/option"
	moblabpb "google.golang.org/genproto/googleapis/chromeos/moblab/v1beta1"

	"infra/cmdsupport/cmdlib"
	"infra/cros/satlab/satlab/internal/commands"
	"infra/cros/satlab/satlab/internal/pkg/google.golang.org/google/chromeos/moblab"
	"infra/cros/satlab/satlab/internal/site"
)

// RunCmd is the implementation of the "satlab run" command.
var RunCmd = &subcommands.Command{
	UsageLine: "run [options...]",
	ShortDesc: "execute a test or suite",
	CommandRun: func() subcommands.CommandRun {
		c := &run{}
		registerRunFlags(c)
		return c
	},
}

// run holds the arguments that are needed for the run command.
type run struct {
	runFlags
}

// Run attempts to run a test or suite and returns an exit status.
func (c *run) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	// Confirm required args are provided and no argument conflicts
	if err := c.validateArgs(); err != nil {
		fmt.Fprintln(a.GetErr(), err.Error())
		c.Flags.Usage()
		cmdlib.PrintError(a, err)
		return 1
	}

	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// InnerRun is the implementation of the run command.
func (c *run) innerRun(a subcommands.Application, positionalArgs []string, env subcommands.Env) error {
	ctx := context.Background()

	// Create TestPlan for suite or test
	var tp *test_platform.Request_TestPlan
	var err error
	if c.suite != "" {
		tp = builder.TestPlanForSuites([]string{c.suite})
	} else if c.test != "" {
		tp = builder.TestPlanForTests(c.testArgs, c.harness, []string{c.test})
	} else if c.testplan != "" {
		fmt.Printf("Fetching testplan...\n")
		var w bytes.Buffer
		err = downloadTestPlan(&w, site.GetGCSImageBucket(), c.testplan)
		if err != nil {
			return err
		}
		tp, err = readTestPlan(c.testplan)
		if err != nil {
			return err
		}
	} else if c.testplan_local != "" {
		tp, err = readTestPlan(c.testplan_local)
		if err != nil {
			return err
		}
		fmt.Printf("Running local testplan...\n")
	}

	dims := c.addedDims
	// Will be nil if not provided by user.
	if dims == nil {
		dims = make(map[string]string)
	}

	// Set drone target based on user input, defaulting to the current box.
	droneDim, err := c.setDroneTarget()
	if err != nil {
		return err
	}
	dims["drone"] = droneDim

	builderId := &buildbucketpb.BuilderID{
		Project: site.GetLUCIProject(),
		Bucket:  site.GetCTPBucket(),
		Builder: site.GetCTPBuilder(),
	}

	if c.image == "" {
		c.image = fmt.Sprintf("%s-release/R%s-%s", c.board, c.milestone, c.build)
	}
	bbClient := &builder.CTPBuilder{
		Image:       c.image,
		Board:       c.board,
		Model:       c.model,
		Pool:        c.pool,
		CFT:         c.cft,
		TestPlan:    tp,
		BuilderID:   builderId,
		Dimensions:  dims,
		ImageBucket: site.GetGCSImageBucket(),
		AuthOptions: &site.DefaultAuthOptions,
		// TRV2:        true,
	}
	// Create default client
	err = bbClient.AddDefaultBBClient(ctx)
	if err != nil {
		return err
	}

	moblabClient, err := moblab.NewBuildClient(ctx, option.WithCredentialsFile(site.GetServiceAccountPath()))
	if err != nil {
		return errors.Annotate(err, "satlab new moblab api build client").Err()
	}

	err = c.innerRunWithClients(ctx, moblabClient, bbClient, site.GetGCSImageBucket())
	if err != nil {
		return errors.Annotate(err, "innerRunWithClients").Err()
	}
	return nil
}

func (c *run) innerRunWithClients(ctx context.Context, moblabClient MoblabClient, bbClient BuildbucketClient, gcsBucket string) error {

	_, _ = c.StageImageToBucket(ctx, moblabClient, gcsBucket)

	_, err := ScheduleBuild(ctx, bbClient)
	if err != nil {
		return errors.Annotate(err, "satlab schedule build").Err()
	}
	return nil
}

func (c *run) StageImageToBucket(ctx context.Context, moblabClient MoblabClient, bucket string) (string, error) {
	if bucket == "" {
		fmt.Println("GCS_BUCKET not found")
		return "", errors.New("GCS_BUCKET not found")
	}
	if c.model == "" {
		c.model = "~"
	}
	if c.image != "" && c.build == "" {
		c.build = strings.Split(c.image, "/")[1]
		c.build = strings.Split(c.build, "-")[1]
	}
	buildTarget := buildTargetParent(c.board, c.model)
	artifactName := fmt.Sprintf("%s/builds/%s/artifacts/%s", buildTarget, c.build, bucket)
	stageReq := &moblabpb.StageBuildRequest{
		Name: artifactName,
	}

	_, err := moblabClient.StageBuild(ctx, stageReq)
	if err != nil {
		return "", err
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
			return "", err
		}
		if stageStatus.IsBuildStaged {
			break
		}
		if count == 0 {
			return "", fmt.Errorf("stage not completed within 10 retries")
		}
	}
	destPath := stageStatus.StagedBuildArtifact.Path

	fmt.Printf("Artifacts staged to %s", path.Join(bucket, destPath))
	return destPath, nil
}

func ScheduleBuild(ctx context.Context, bbClient BuildbucketClient) (string, error) {
	ctpBuild, err := bbClient.ScheduleCTPBuild(ctx)
	if err != nil {
		return "", err
	}
	fmt.Printf("\n\n-- BUILD LINK --\nhttps://ci.chromium.org/ui/b/%s\n\n", strconv.Itoa(int(ctpBuild.Id)))
	return "", nil
}

func (c *run) validateArgs() error {
	executionTarget := 0
	if c.testplan != "" {
		executionTarget++
	}
	if c.testplan_local != "" {
		executionTarget++
	}
	if c.suite != "" {
		executionTarget++
	}
	if c.test != "" {
		executionTarget++
	}
	if executionTarget != 1 {
		return errors.Reason("Please specify only one of the following: -suite, -test, -testplan, -testplan_local").Err()
	}
	if c.cft && c.test != "" && c.harness == "" {
		return errors.Reason("-harness is required for cft test runs").Err()
	}
	if c.board == "" {
		return errors.Reason("-board not specified").Err()
	}
	if c.image == "" {
		if c.model == "" {
			return errors.Reason("-model must be specified if -image is not provided").Err()
		}
		if c.milestone == "" {
			return errors.Reason("-milestone must be specified if -image is not provided").Err()
		}
		if c.build == "" {
			return errors.Reason("-build must be specified if -image is not provided").Err()
		}
	}
	if c.pool == "" {
		return errors.Reason("-pool not specified").Err()
	}
	if _, ok := c.addedDims["drone"]; ok {
		return errors.Reason("-dims cannot include drone (control via -satlabId instead)").Err()
	}
	return nil
}

// Set drone target to user-provided satlab or local satlab if one isn't provided
func (c *run) setDroneTarget() (string, error) {
	var satlabTarget string
	if c.satlabId != "" {
		satlabTarget = fmt.Sprintf(c.satlabId)
	} else { // get id of local satlab if one is not provided
		localSatlab, err := commands.GetDockerHostBoxIdentifier()
		if err != nil {
			return "", errors.Annotate(err, "satlab get docker host box identifier").Err()
		}
		satlabTarget = fmt.Sprintf("satlab-%s", localSatlab)
	}
	return satlabTarget, nil
}

func buildTargetParent(board string, model string) string {
	artifactName := fmt.Sprintf("buildTargets/%s/models/%s", board, model)
	return artifactName
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
func downloadTestPlan(w io.Writer, bucket, testplan string) error {
	object := "testplans/" + testplan
	destFileName := "/config/" + testplan
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(site.GetServiceAccountPath()))
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("%q Error: %w", object, err)
	}
	defer rc.Close()

	f, err := os.Create(destFileName)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	fmt.Fprintf(w, "Blob %v downloaded to local file %v\n", object, destFileName)

	return nil
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
