// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package build_service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	moblabapipb "google.golang.org/genproto/googleapis/chromeos/moblab/v1beta1"
	"google.golang.org/genproto/protobuf/field_mask"

	"infra/cros/recovery/models"
	moblabapi "infra/cros/satlab/common/google.golang.org/google/chromeos/moblab"
	"infra/cros/satlab/common/utils/collection"
	"infra/cros/satlab/common/utils/parser"
)

// PageSize The number of items to return in a page
const PageSize = 1000

// ParseBuildTargetsPath compose the path by given board.
func ParseBuildTargetsPath(board string) string {
	// TODO find the better way to do
	return fmt.Sprintf("buildTargets/%s", board)
}

// ParseModelPath compose the path by given board and model.
func ParseModelPath(board string, model string) string {
	// TODO find the better way to do
	return fmt.Sprintf("buildTargets/%s/models/%s", board, model)
}

// ParseBuildArtifactPath compose the path by given board, model, buildVersion, and bucket.
func ParseBuildArtifactPath(board string, model string, buildVersion string, bucket string) string {
	// TODO find the better way to do
	return fmt.Sprintf("buildTargets/%s/models/%s/builds/%s/artifacts/%s", board, model, buildVersion, bucket)
}

// BuildServiceImpl is an object for connecting the build client.
type BuildServiceImpl struct {
	// client the `BuildClient`
	client *moblabapi.BuildClient
}

// New sets up the `BuildClient` and returns a BuildConnector.
// The service account is set in the global environment.
func New(ctx context.Context) (IBuildService, error) {
	// Set your service account: $ export GOOGLE_APPLICATION_CREDENTIALS="service_account.json"
	// Client need not be created for each request
	client, err := moblabapi.NewBuildClient(ctx)
	if err != nil {
		return nil, err
	}
	return &BuildServiceImpl{
		client: client,
	}, nil
}

// ListBuildTargets returns all the board.
func (b *BuildServiceImpl) ListBuildTargets(ctx context.Context) ([]string, error) {
	log.Println("Trying to list build targets")

	req := &moblabapipb.ListBuildTargetsRequest{
		PageSize: PageSize,
	}

	iter := b.client.ListBuildTargets(ctx, req)
	res, err := collection.Collect(
		iter.Next,
		func(board *moblabapipb.BuildTarget) (string, error) {
			return board.GetName(), nil
		},
	)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// ListModels returns all models by given board.
//
// string board is the board name that we use it as a filter.
func (b *BuildServiceImpl) ListModels(ctx context.Context, board string) ([]string, error) {
	log.Println("Trying to list models")

	parent := ParseBuildTargetsPath(board)

	req := &moblabapipb.ListModelsRequest{
		Parent:   parent,
		PageSize: PageSize,
	}

	iter := b.client.ListModels(ctx, req)

	res, err := collection.Collect(
		iter.Next,
		func(model *moblabapipb.Model) (string, error) {
			return model.GetName(), nil
		},
	)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// ListAvailableMilestones returns all available milestones by given board and model.
//
// string board is the board name that we use it as a filter.
// string model is the model name that we use it as a filter.
func (b *BuildServiceImpl) ListAvailableMilestones(ctx context.Context, board string, model string) ([]string, error) {
	log.Println("Trying to list available milestones")

	fm := &field_mask.FieldMask{
		Paths: []string{"milestone"},
	}

	req := &moblabapipb.ListBuildsRequest{
		Parent:   ParseModelPath(board, model),
		ReadMask: fm,
		GroupBy:  fm,
		PageSize: PageSize,
	}

	iter := b.client.ListBuilds(ctx, req)

	res, err := collection.Collect(
		iter.Next,
		func(build *moblabapipb.Build) (string, error) {
			milestone, err := parser.ExtractMilestoneFrom(build.GetMilestone())
			if err != nil {
				log.Printf("the milestone format isn't match %v\n", build.GetMilestone())
				return "", err
			}
			return milestone, nil
		},
	)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (b *BuildServiceImpl) findMostStableBuildByBoard(ctx context.Context, board string) (*moblabapipb.Build, error) {
	buildTarget := ParseBuildTargetsPath(board)

	req := &moblabapipb.FindMostStableBuildRequest{
		BuildTarget: buildTarget,
	}

	resp, err := b.client.FindMostStableBuild(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetBuild(), nil
}

func buildToOS(milestone, build string) string {
	return fmt.Sprintf("R%s-%s", milestone, build)
}

// FindMostStableBuild find the stable build version by given board.
//
// string board is the board name that we use it as a filter.
func (b *BuildServiceImpl) FindMostStableBuild(ctx context.Context, board string) (string, error) {
	resp, err := b.findMostStableBuildByBoard(ctx, board)
	if err != nil {
		return "", err
	}

	milestone, err := parser.ExtractMilestoneFrom(resp.GetMilestone())
	if err != nil {
		return "", errors.New(fmt.Sprintf("milestone pattern doesn't match %v\n", resp.GetMilestone()))
	}

	return buildToOS(milestone, resp.GetBuildVersion()), nil
}

// FindMostStableBuildByBoardAndModel find the stable recovery version by board and model
func (b *BuildServiceImpl) FindMostStableBuildByBoardAndModel(ctx context.Context, board, model string) (*models.RecoveryVersion, error) {
	resp, err := b.findMostStableBuildByBoard(ctx, board)
	if err != nil {
		return nil, err
	}
	milestone, err := parser.ExtractMilestoneFrom(resp.GetMilestone())
	os := buildToOS(milestone, resp.GetBuildVersion())
	fw := resp.GetRwFirmwareVersion()

	listMilestonesRequest := &moblabapipb.ListBuildsRequest{
		Parent: ParseModelPath(board, model),
		Filter: "type=firmware",
	}
	listMilestonesResponse := b.client.ListBuilds(ctx, listMilestonesRequest)
	milestoneBuild, err := listMilestonesResponse.Next()
	if err != nil {
		return nil, err
	}
	fwMilestoneList := strings.Split(milestoneBuild.GetMilestone(), "/")
	if len(fwMilestoneList) < 2 {
		return nil, errors.New("Invalid milestone")
	}
	fwMilestone := fwMilestoneList[1]

	// fetch firmware build version
	listBuildRequest := &moblabapipb.ListBuildsRequest{
		Parent:   ParseModelPath(board, model),
		Filter:   fmt.Sprintf("type=firmware+milestone=milestones/%s", fwMilestone),
		PageSize: 1,
	}
	listBuildResponse := b.client.ListBuilds(ctx, listBuildRequest)
	firmwareBuild, err := listBuildResponse.Next()
	if err != nil {
		return nil, err
	}
	fwImage := fmt.Sprintf("%s-firmware/R%s-%s", board, fwMilestone, firmwareBuild.GetBuildVersion())

	return &models.RecoveryVersion{
		Board:     board,
		Model:     model,
		OsImage:   os,
		FwVersion: fw,
		FwImage:   fwImage,
	}, nil

}

// ListBuildsForMilestone returns all build versions by given board, model, and milestone.
//
// string board is the board name that we use it as a filter.
// string model is the model name that we use it as a filter.
// int32 milestone is the milestone that we use it as a filter.
func (b *BuildServiceImpl) ListBuildsForMilestone(
	ctx context.Context,
	board string,
	model string,
	milestone int32,
) ([]*BuildVersion, error) {
	filter := fmt.Sprintf("milestone=milestones/%d", milestone)
	req := &moblabapipb.ListBuildsRequest{
		Parent:   ParseModelPath(board, model),
		Filter:   filter,
		PageSize: PageSize,
	}

	iter := b.client.ListBuilds(ctx, req)

	res, err := collection.Collect(
		iter.Next,
		func(build *moblabapipb.Build) (*BuildVersion, error) {
			status := FromGCSBucketBuildStatusMap[build.GetStatus()]
			return &BuildVersion{
				Version: build.GetBuildVersion(),
				Status:  status,
			}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// CheckBuildStageStatus check the build version is staged by given board, model, build version, and bucket name.
//
// string board is the board name that we use it as a filter.
// string model is the model name that we use it as a filter.
// string buildVersion is the build version that we use it as a filter.
// string bucketName the bucket we need to check the build version is in this bucket.\fc
func (b *BuildServiceImpl) CheckBuildStageStatus(
	ctx context.Context,
	board string,
	model string,
	buildVersion string,
	bucketName string,
) (bool, error) {
	req := &moblabapipb.CheckBuildStageStatusRequest{
		Name: ParseBuildArtifactPath(board, model, buildVersion, bucketName),
	}

	res, err := b.client.CheckBuildStageStatus(ctx, req)
	if err != nil {
		return false, err
	}

	return res.IsBuildStaged, nil
}

// StageBuild stage the build version in the bucket by given board, model, build version, and bucket name.
//
// string board is the board that we want to stage.
// string model is the model that we want to stage.
// string buildVersion is the build version that we want to stage.
// string bucketName which bucket we want to put the build version in.
func (b *BuildServiceImpl) StageBuild(ctx context.Context,
	board string,
	model string,
	buildVersion string,
	bucketName string,
) (*moblabapipb.BuildArtifact, error) {
	req := &moblabapipb.StageBuildRequest{
		Name: ParseBuildArtifactPath(board, model, buildVersion, bucketName),
	}

	operation, err := b.client.StageBuild(ctx, req)
	if err != nil {
		return nil, err
	}

	res, err := operation.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return res.GetStagedBuildArtifact(), nil
}

// Close to close the client connection.
func (b *BuildServiceImpl) Close() error {
	return b.client.Close()
}
