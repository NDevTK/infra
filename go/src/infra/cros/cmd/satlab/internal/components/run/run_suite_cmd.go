// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"context"
	"fmt"
	"path"
	"strconv"

	"github.com/googleapis/gax-go/v2"
	"github.com/maruel/subcommands"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/chromiumos/platform/dev-util/src/chromiumos/ctp/builder"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/api/option"
	moblabpb "google.golang.org/genproto/googleapis/chromeos/moblab/v1beta1"

	"infra/cmdsupport/cmdlib"
	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/pkg/google.golang.org/google/chromeos/moblab"
	"infra/cros/cmd/satlab/internal/site"
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
	if c.suite != "" {
		tp = builder.TestPlanForSuites([]string{c.suite})
	} else {
		tp = builder.TestPlanForTests(c.testArgs, c.harness, []string{c.test})
	}

	// Set drone target based on user input
	drone, err := c.setDroneTarget()
	if err != nil {
		return err
	}
	dims := map[string]string{"drone": drone}

	builderId := &buildbucketpb.BuilderID{
		Project: site.LUCIProject,
		Bucket:  site.BuilderBucket,
		Builder: site.CTPBuilder,
	}

	bbClient := &builder.CTPBuilder{
		Image:       fmt.Sprintf("%s-release/R%s-%s", c.board, c.milestone, c.build),
		Board:       c.board,
		Model:       c.model,
		Pool:        c.pool,
		CFT:         true,
		TestPlan:    tp,
		BuilderID:   builderId,
		Dimensions:  dims,
		ImageBucket: site.GCSBucket,
		AuthOptions: &site.DefaultAuthOptions,
		// TRV2:        true,
	}
	// Create default client
	err = bbClient.AddDefaultBBClient(ctx)
	if err != nil {
		return err
	}

	moblabClient, err := moblab.NewBuildClient(ctx, option.WithCredentialsFile("/home/satlab/keys/service_account.json"))
	if err != nil {
		return errors.Annotate(err, "satlab new moblab api build client").Err()
	}

	err = c.innerRunWithClients(ctx, moblabClient, bbClient, site.GCSBucket)
	if err != nil {
		return errors.Annotate(err, "innerRunWithClients").Err()
	}
	return nil
}

func (c *run) innerRunWithClients(ctx context.Context, moblabClient MoblabClient, bbClient BuildbucketClient, gcsBucket string) error {
	_, err := c.StageImageToBucket(ctx, moblabClient, gcsBucket)
	if err != nil {
		return errors.Annotate(err, "satlab stage image to bucket").Err()
	}
	_, err = ScheduleBuild(ctx, bbClient)
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
	if (c.suite == "" && c.test == "") || (c.suite != "" && c.test != "") {
		return errors.Reason("Please specify either -suite or -test").Err()
	}
	if c.test != "" && c.harness == "" {
		return errors.Reason("-harness is required for individual test execution").Err()
	}
	if c.board == "" {
		return errors.Reason("-board not specified").Err()
	}
	if c.model == "" {
		return errors.Reason("-model not specified").Err()
	}
	if c.milestone == "" {
		return errors.Reason("-milestone not specified").Err()
	}
	if c.build == "" {
		return errors.Reason("-build not specified").Err()
	}
	// Set default name of a pool if no information is given
	if c.pool == "" {
		c.pool = "xolabs-satlab"
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
