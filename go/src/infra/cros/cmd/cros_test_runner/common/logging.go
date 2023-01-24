// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"fmt"
	"os"
	"path"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// WriteProtoToStepLog writes provided proto to build step.
func WriteProtoToStepLog(ctx context.Context, step *build.Step, proto proto.Message, logText string) {
	outputLog := step.Log(logText)
	marshaller := protojson.MarshalOptions{Multiline: true, Indent: "  "}
	bytes, err := marshaller.Marshal(proto)
	if err != nil {
		logging.Infof(ctx, "%s: %q", fmt.Sprintf("Marshalling %q failed:", logText), err.Error())
	}
	_, err = outputLog.Write(bytes)
	if err != nil {
		logging.Infof(ctx, "%s: %q", fmt.Sprintf("Writing %q failed:", logText), err.Error())
	}
}

// WriteProtoToJsonFile writes provided proto to a json file.
func WriteProtoToJsonFile(ctx context.Context, dirPath string, fileName string, inputProto proto.Message) (string, error) {
	protoFilePath := path.Join(dirPath, fileName)
	f, err := os.Create(protoFilePath)
	if err != nil {
		return "", fmt.Errorf("error during creating file %q: %s", fileName, err.Error())
	}
	defer f.Close()

	bytes, err := protojson.Marshal(inputProto)
	if err != nil {
		return "", fmt.Errorf("error during marshalling proto for %q: %s", fileName, err.Error())
	}

	_, err = f.Write(bytes)
	if err != nil {
		return "", fmt.Errorf("error during writing proto to file %q: %s", fileName, err.Error())
	}

	logging.Infof(ctx, "proto successfully written to file: %s", bytes)

	err = f.Close()
	if err != nil {
		return "", fmt.Errorf("error during closing file %q: %s", fileName, err.Error())
	}

	return protoFilePath, nil
}

// LogExecutionDetails logs provided cmds to build step.
func LogExecutionDetails(ctx context.Context, step *build.Step, cmds []string) {
	outputLog := step.Log("Execution Details")
	logCmdsStr := fmt.Sprintf("%+q", cmds)
	_, err := outputLog.Write([]byte(logCmdsStr))
	if err != nil {
		logging.Infof(ctx, "Failed to create execution details for cmd: %q", cmds)
	}
}
