// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package stableversion

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/googleapis/gax-go/v2"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"
	"google.golang.org/api/option"
	moblabpb "google.golang.org/genproto/googleapis/chromeos/moblab/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	"infra/cros/recovery/models"
	"infra/cros/satlab/common/google.golang.org/google/chromeos/moblab"
	"infra/cros/satlab/common/run"
	"infra/cros/satlab/common/site"
)

// If allowSetModelBoard is true, then the user is allowed to create new entries for a host&model.
// If allowSetModelBoard is false, then the user is blocked from creating new entries a host&model.
//
// We set this variable to false. We want to force users to use the per-host override so that it is
// easier to replace the stable version implementation with a new service behind the scenes without
// a change in user-facing behavior.
const allowSetModelBoard = false

var SetStableVersionCmd = &subcommands.Command{
	UsageLine: `set-stable-version`,
	ShortDesc: `Set the stable version using {board, model} or {hostname}.`,
	CommandRun: func() subcommands.CommandRun {
		r := &setStableVersionRun{}

		r.authFlags.Register(&r.Flags, site.DefaultAuthOptions)
		r.envFlags.Register(&r.Flags)
		r.commonFlags.Register(&r.Flags)

		// if allowSetModelBoard (
		r.Flags.StringVar(&r.board, "board", "", `the board or build target (used with model)`)
		r.Flags.StringVar(&r.model, "model", "", `the model (used with board)`)
		// )

		r.Flags.StringVar(&r.hostname, "hostname", "", `the hostname (used by itself)`)
		r.Flags.StringVar(&r.os, "os", "", `the OS version to set (no change if empty)`)
		r.Flags.StringVar(&r.fw, "fw", "", `the firmware version to set (no change if empty)`)
		r.Flags.StringVar(&r.fwImage, "fwImage", "", `the firmware image version to set (no change if empty)`)
		return r
	},
}

// SetStableVersionRun is the command for adminclient set-stable-version.
type setStableVersionRun struct {
	subcommands.CommandRunBase

	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	board    string
	model    string
	hostname string
	os       string
	fw       string
	fwImage  string
}

// Run runs the command, prints the error if there is one, and returns an exit status.
func (c *setStableVersionRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, c, env)
	if err := c.innerRun(ctx, a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

// InnerRun implements Run using Board/Model or Hostname depending on provided args.
func (c *setStableVersionRun) innerRun(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
	ctx = utils.SetupContext(ctx, c.envFlags.GetNamespace())
	err := c.validateArgs()
	if err != nil {
		return err
	}

	// Hostname flow: INTERNAL USERS ONLY
	if c.hostname != "" {
		err := c.innerRunHostname(ctx, a, args, env)
		if err != nil {
			return err
		}
	} else if c.board != "" && c.model != "" { // Board Model flow: SFP EXTERNAL USERS ONLY
		recovery_version := &models.RecoveryVersion{
			Board:     c.board,
			Model:     c.model,
			OsImage:   c.os,
			FwVersion: c.fw,
			FwImage:   c.fwImage,
		}
		err := innerRunBoardModel(ctx, a, args, env, recovery_version)
		if err != nil {
			return err
		}
	}
	return nil
}

// InnerRunBoardModel is the implementation of setStableVersion that uses Board/Model and circumvents the cros-inventory call.
func innerRunBoardModel(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env, rv *models.RecoveryVersion) error {

	fmt.Println("WARNING: internal users must use -hostname override instead of board/model")

	// Count how many of OS, FW or FW Image are provided
	numArgs := validateStableVersionArgs(rv.OsImage, rv.FwVersion, rv.FwImage)
	moblabClient, err := moblab.NewBuildClient(ctx, option.WithCredentialsFile(site.GetServiceAccountPath()))
	if numArgs == 0 { // If none provided, use board/model to fetch arbitrary version
		if err != nil {
			return errors.Annotate(err, "satlab new moblab api build client").Err()
		}
		rv, err = FindMostStableBuild(ctx, moblabClient, rv.Board, rv.Model)
		if err != nil {
			return errors.Annotate(err, "find most stable build").Err()
		}
	} else if numArgs < 3 { // If partial args provided, throw an error
		return fmt.Errorf("Please provide all or none of the following: -os, -fw, -fwImage")
	}
	err = StageAndWriteLocalStableVersion(ctx, moblabClient, rv)
	if err != nil {
		return errors.Annotate(err, "stage and write local stable version").Err()
	}
	return nil
}

// InnerRunHostname is the implementation of setStableVersion for internal Satlab users that requires hostname et al.
func (c *setStableVersionRun) innerRunHostname(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {

	fmt.Println("WARNING: Satlab for Partners users must use board/model override instead of hostname")

	newHostname, err := preprocessHostname(ctx, c.commonFlags, c.hostname, nil, nil)
	if err != nil {
		return errors.Annotate(err, "set stable version").Err()
	}
	c.hostname = newHostname

	req, err := c.produceRequest(ctx, a, args, env)
	if err != nil {
		return errors.Annotate(err, "set stable version").Err()
	}

	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return errors.Annotate(err, "set stable version").Err()
	}

	invWithSVClient := fleet.NewInventoryPRPCClient(
		&prpc.Client{
			C:       hc,
			Host:    c.envFlags.GetCrosAdmService(),
			Options: site.DefaultPRPCOptions,
		},
	)

	resp, err := invWithSVClient.SetSatlabStableVersion(ctx, req)
	if err != nil {
		return errors.Annotate(err, "get stable version").Err()
	}
	out, err := protojson.MarshalOptions{
		Indent: "  ",
	}.Marshal(resp)
	if err != nil {
		return errors.Annotate(err, "get stable version").Err()
	}
	fmt.Fprintf(a.GetOut(), "%s\n", out)
	return nil
}

// StageAndWriteLocalStableVersion stages a recovery image to partner bucket and writes the associated rv metadata locally
func StageAndWriteLocalStableVersion(ctx context.Context, moblabClient MoblabClient, rv *models.RecoveryVersion) error {
	buildVersion := strings.Split(rv.OsImage, "-")[1]
	err := run.StageImageToBucket(ctx, moblabClient, rv.Board, rv.Model, buildVersion)
	if err != nil {
		return errors.Annotate(err, "stage stable version image to bucket").Err()
	}
	err = writeLocalStableVersion(rv, site.RecoveryVersionDirectory)
	if err != nil {
		return errors.Annotate(err, "write local stable version").Err()
	}
	return nil
}

// WriteLocalStableVersion saves a recovery version to the specified directory and creates the directory if necessary.
func writeLocalStableVersion(recovery_version *models.RecoveryVersion, path string) error {

	// Check if recovery_versions directory created
	_, err := os.Stat(path)
	if err != nil {
		return err
	}

	fname := fmt.Sprintf("%s%s-%s.json", path, recovery_version.Board, recovery_version.Model)
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	// close file on exit and check for its returned error
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	rv, err := json.MarshalIndent(recovery_version, "", " ")
	if err != nil {
		return errors.Annotate(err, "marshal recovery version").Err()
	}
	_, err = f.Write(rv)
	if err != nil {
		return err
	}
	fmt.Println("Recovery Version written locally: ", string(rv))

	return nil
}

// Fetch a stable recovery version for a given board model
func FindMostStableBuild(ctx context.Context, moblabClient MoblabClient, board string, model string) (*models.RecoveryVersion, error) {

	// fetch os image and fw
	findBuildRequest := &moblabpb.FindMostStableBuildRequest{
		BuildTarget: "buildTargets/" + board,
	}
	resp, err := moblabClient.FindMostStableBuild(ctx, findBuildRequest)
	if err != nil {
		return nil, err
	}
	milestone := strings.Split(resp.GetBuild().GetMilestone(), "/")[1]
	os := "R" + milestone + "-" + resp.Build.GetBuildVersion()
	fw := resp.Build.GetRwFirmwareVersion()

	listMilestonesRequest := &moblabpb.ListBuildsRequest{
		Parent: fmt.Sprintf("buildTargets/%s/models/%s", board, model),
		Filter: "type=firmware",
	}
	listMilestonesResponse := moblabClient.ListBuilds(ctx, listMilestonesRequest)
	milestoneBuild, err := listMilestonesResponse.Next()
	if err != nil {
		return nil, err
	}
	fw_milestone := strings.Split(milestoneBuild.GetMilestone(), "/")[1]

	// fetch firmware build version
	listBuildVersionsRequest := &moblabpb.ListBuildsRequest{
		Parent:   fmt.Sprintf("buildTargets/%s/models/%s", board, model),
		Filter:   fmt.Sprintf("type=firmware+milestone=milestones/%s", fw_milestone),
		PageSize: 1,
	}
	listBuildVersionsResponse := moblabClient.ListBuilds(ctx, listBuildVersionsRequest)
	firmwareBuild, err := listBuildVersionsResponse.Next()
	if err != nil {
		return nil, err
	}
	fwImage := fmt.Sprintf("%s-firmware/R%s-%s", board, fw_milestone, firmwareBuild.GetBuildVersion())

	rv := &models.RecoveryVersion{
		Board:     board,
		Model:     model,
		OsImage:   os,
		FwVersion: fw,
		FwImage:   fwImage,
	}
	return rv, nil
}

func (c *setStableVersionRun) validateArgs() error {
	if (c.board != "" && c.model == "") || (c.board == "" && c.model != "") {
		return errors.Reason("Please provide both -board AND -model").Err()
	}
	if c.hostname == "" && c.board == "" && c.model == "" {
		return errors.Reason("Please provide either -hostname OR -board and -model").Err()
	}
	return nil
}

func validateStableVersionArgs(os string, fw string, fwImage string) int {
	count := 0
	if os != "" {
		count++
	}
	if fw != "" {
		count++
	}
	if fwImage != "" {
		count++
	}
	return count
}

// ProduceRequest creates a request that can be used as a key to set the stable version.
// If the command line arguments do not unambiguously indicate how to create such a request, we fail.
func (c *setStableVersionRun) produceRequest(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) (*fleet.SetSatlabStableVersionRequest, error) {
	req := &fleet.SetSatlabStableVersionRequest{}
	useHostnameStrategy := c.hostname != ""
	useBoardModelStrategy := allowSetModelBoard && (c.board != "") && (c.model != "")

	// Validate and populate the strategy field of the request.
	if err := func() error {
		if useHostnameStrategy {
			if useBoardModelStrategy {
				return errors.Reason("board and model should not be set if hostname is provided").Err()
			}
			req.Strategy = &fleet.SetSatlabStableVersionRequest_SatlabHostnameStrategy{
				SatlabHostnameStrategy: &fleet.SatlabHostnameStrategy{
					Hostname: c.hostname,
				},
			}
			return nil
		} // Hostname strategy not set.
		if !useBoardModelStrategy {
			return errors.Reason("must provide hostname").Err()
		}
		req.Strategy = &fleet.SetSatlabStableVersionRequest_SatlabBoardAndModelStrategy{
			SatlabBoardAndModelStrategy: &fleet.SatlabBoardAndModelStrategy{
				Board: c.board,
				Model: c.model,
			},
		}
		return nil
	}(); err != nil {
		return nil, err
	}

	// TODO(gregorynisbet): Consider adding validation here instead of on the server side
	req.CrosVersion = c.os
	req.FirmwareVersion = c.fw
	req.FirmwareImage = c.fwImage

	return req, nil
}

// MoblabClient interface provides subset of Moblab API methods relevant to Stable Version functionality
type MoblabClient interface {
	FindMostStableBuild(ctx context.Context, req *moblabpb.FindMostStableBuildRequest, opts ...gax.CallOption) (*moblabpb.FindMostStableBuildResponse, error)
	ListBuilds(ctx context.Context, req *moblabpb.ListBuildsRequest, opts ...gax.CallOption) *moblab.BuildIterator
	StageBuild(ctx context.Context, req *moblabpb.StageBuildRequest, opts ...gax.CallOption) (*moblab.StageBuildOperation, error)
	CheckBuildStageStatus(ctx context.Context, req *moblabpb.CheckBuildStageStatusRequest, opts ...gax.CallOption) (*moblabpb.CheckBuildStageStatusResponse, error)
}
