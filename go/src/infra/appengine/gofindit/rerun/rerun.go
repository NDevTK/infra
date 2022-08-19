// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package rerun handles rerun for a build.
package rerun

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"

	"infra/appengine/gofindit/internal/buildbucket"
)

// TriggerRerun triggers a rerun build for a particular build bucket build and Gitiles commit.
func TriggerRerun(c context.Context, commit *buildbucketpb.GitilesCommit, failedBuildID int64, analysisID int64) (*buildbucketpb.Build, error) {
	logging.Infof(c, "triggerRerun with commit %s", commit.Id)
	properties, dimensions, err := getRerunPropertiesAndDimensions(c, failedBuildID, analysisID)
	if err != nil {
		logging.Errorf(c, "Failed getRerunPropertiesAndDimension for build %d", failedBuildID)
		return nil, err
	}
	req := &buildbucketpb.ScheduleBuildRequest{
		Builder: &buildbucketpb.BuilderID{
			Project: "chromium",
			Bucket:  "findit",
			Builder: "gofindit-culprit-verification",
		},
		Properties:    properties,
		Dimensions:    dimensions,
		Tags:          getRerunTags(c, failedBuildID),
		GitilesCommit: commit,
	}
	build, err := buildbucket.ScheduleBuild(c, req)
	if err != nil {
		logging.Errorf(c, "Failed trigger rerun for build %d: %w", failedBuildID, err)
		return nil, err
	}
	logging.Infof(c, "Rerun build %d triggered for build: %d", build.GetId(), failedBuildID)
	return build, nil
}

// getRerunTags returns the build bucket tags for the rerun build
func getRerunTags(c context.Context, bbid int64) []*buildbucketpb.StringPair {
	return []*buildbucketpb.StringPair{
		{
			// analyzed_build_id is the buildbucket ID of the build which we want to rerun.
			Key:   "analyzed_build_id",
			Value: strconv.FormatInt(bbid, 10),
		},
	}
}

// getRerunProperty returns the properties and dimensions for a rerun of a buildID
func getRerunPropertiesAndDimensions(c context.Context, bbid int64, analysisID int64) (*structpb.Struct, []*buildbucketpb.RequestedDimension, error) {
	build, err := buildbucket.GetBuild(c, bbid, &buildbucketpb.BuildMask{
		Fields: &fieldmaskpb.FieldMask{
			Paths: []string{"input.properties", "builder", "infra.swarming.task_dimensions"},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get properties for build %d: %w", bbid, err)
	}
	properties, err := getRerunProperties(c, build, analysisID)
	if err != nil {
		return nil, nil, err
	}
	dimens := getRerunDimensions(c, build)
	return properties, dimens, nil
}

func getRerunProperties(c context.Context, build *buildbucketpb.Build, analysisID int64) (*structpb.Struct, error) {
	fields := map[string]interface{}{}
	properties := build.GetInput().GetProperties()
	if properties != nil {
		m := properties.GetFields()
		if builderGroup, ok := m["builder_group"]; ok {
			fields["builder_group"] = builderGroup
			fields["target_builder"] = map[string]string{
				"builder": build.Builder.Builder,
				"group":   builderGroup.GetStringValue(),
			}
		}
		if bootstrapProperties, ok := m["$bootstrap/properties"]; ok {
			fields["$bootstrap/properties"] = bootstrapProperties
		}
	}
	fields["analysis_id"] = analysisID
	spb, err := toStructPB(fields)
	if err != nil {
		return nil, fmt.Errorf("cannot convert %v to structpb: %w", fields, err)
	}
	return spb, nil
}

func getRerunDimensions(c context.Context, build *buildbucketpb.Build) []*buildbucketpb.RequestedDimension {
	result := []*buildbucketpb.RequestedDimension{}

	// Only copy these dimensions from the analyzed builder to the rerun job request.
	allowedDimensions := map[string]bool{"os": true, "gpu": true}
	if build.GetInfra() != nil && build.GetInfra().GetSwarming() != nil && build.GetInfra().GetSwarming().GetTaskDimensions() != nil {
		dimens := build.GetInfra().GetSwarming().GetTaskDimensions()
		for _, d := range dimens {
			if _, ok := allowedDimensions[d.Key]; ok {
				result = append(result, &buildbucketpb.RequestedDimension{
					Key:   d.Key,
					Value: d.Value,
				})
			}
		}
	}
	return result
}

// TODO (nqmtuan): Move this into a helper class if it turns out we need to use
// it for more than one place
// toStructPB convert an interface{} s to structpb.Struct, as long as s is marshallable.
// s can be a general Go type, structpb.Struct type, or mixed.
// For example, s can be a map of mixed type, like
// {"key1": "val1", "key2": structpb.NewStringValue("val2")}
func toStructPB(s interface{}) (*structpb.Struct, error) {
	// We used json as an intermediate format to convert
	j, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(j, &m); err != nil {
		return nil, err
	}
	return structpb.NewStruct(m)
}
