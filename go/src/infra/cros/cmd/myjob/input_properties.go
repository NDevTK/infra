// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/led/job"
	"google.golang.org/protobuf/types/known/structpb"
)

func (m myjobRunBase) GetBuilderInputProps(ctx context.Context, fullBuilderName string) (*structpb.Struct, error) {
	bucket, builder, err := separateBucketFromBuilder(fullBuilderName)
	if err != nil {
		return &structpb.Struct{}, err
	}

	stdout, stderr, err := m.RunCmd(ctx, "led", "get-builder", fmt.Sprintf("%s:%s", bucket, builder))
	if err != nil {
		fmt.Println(stderr)
		return &structpb.Struct{}, err
	}

	var definition job.Definition_Buildbucket
	if err := json.Unmarshal([]byte(stdout), &definition); err != nil {
		// `led get-builder` returns proto enum fields as strings instead of ints, like buildbucket.bbagent_args.infra.experiment_reasons.
		// json.Unmarshal considers that to be an inappropriate type, returns an UnmarshalTypeError, and otherwise unmarshals as best as it can.
		// If the error is an UnmarshalTypeError, and InputProperties seem to have been unmarshaled OK, then there's no problem.
		inputProps := definitionToInputProperties(definition)
		if _, ok := err.(*json.UnmarshalTypeError); ok && inputProps != nil {
			return inputProps, nil
		}
		return &structpb.Struct{}, errors.Annotate(err, "unmarshaling `led get-builder` output").Err()
	}
	return definitionToInputProperties(definition), nil
}

// definitionToInputProperties extracts the input properties from an unmarshaled `led get-builder` output.
func definitionToInputProperties(definition job.Definition_Buildbucket) *structpb.Struct {
	return definition.Buildbucket.GetBbagentArgs().GetBuild().GetInput().GetProperties()
}

// writeStructToFile creates a tempfile, writes the struct as JSON data, and returns the File object.
func writeStructToFile(s *structpb.Struct) (*os.File, error) {
	file, err := os.CreateTemp("", "input_props")
	if err != nil {
		return nil, err
	}
	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if _, err := file.Write(jsonBytes); err != nil {
		return nil, err
	}
	return file, nil
}
