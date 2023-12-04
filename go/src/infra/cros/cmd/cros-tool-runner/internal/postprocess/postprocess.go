// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
// Package postprocess to interface with post-process.
package postprocess

import (
	"context"
	"infra/cros/cmd/cros-tool-runner/internal/common"
	"infra/cros/cmd/cros-tool-runner/internal/services"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/golang/protobuf/jsonpb"
	build_api "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
)

const (
	PostProcessName = "post-process"
	CrostDutName    = "cros-dut"

	// Cros-dut result temp dir.
	crosDutResultsTempDir = "cros-dut-results"

	tempDirPath = "/var/tmp"
)

// Run post-process.
func Run(ctx context.Context, req *api.CrosToolRunnerPostTestRequest, crosDutContainer, postProcessContainer *build_api.ContainerImageInfo, tokenFile string) (res *api.CrosToolRunnerPostTestResponse, err error) {
	// Use host network for dev environment which DUT address is in the form localhost:<port>
	const (
		networkName = "host"
	)
	artifactDir, err := filepath.Abs(req.ArtifactDir)
	if err != nil {
		return nil, errors.Annotate(err, "prepare to run postprocess: failed to resolve artifact directory %v", req.ArtifactDir).Err()
	}
	// All artifacts will be in <artifact_dir>/post-process.
	postProcessDir := path.Join(artifactDir, PostProcessName)

	inputFileName := path.Join(postProcessDir, "request.json")

	// Setting up directories.
	if err := os.MkdirAll(postProcessDir, 0755); err != nil {
		return nil, errors.Annotate(err, "prepare to run postprocess: failed to create directory %s", postProcessDir).Err()
	}

	device := req.GetPrimaryDut()
	dutName := device.GetDut().GetId().GetValue()
	dutSshInfo := device.GetDut().GetChromeos().GetSsh()

	cacheServerInfo := device.GetDut().GetCacheServer()

	parentTempDir := ""
	if _, err := os.Stat(tempDirPath); err == nil {
		parentTempDir = tempDirPath
	}

	// Create temp results dir for cros-dut
	crosDutResultsDir, err := ioutil.TempDir(parentTempDir, crosDutResultsTempDir)
	if err != nil {
		log.Printf("cros-dut results temp directory creation failed with error: %s", err)
		return nil, errors.Annotate(err, "create dut service: create temp dir").Err()
	}

	log.Printf("--> Starting cros-dut service for %q ...", dutName)
	dutService, err := services.CreateDutService(ctx, crosDutContainer, dutName, networkName, cacheServerInfo, dutSshInfo, crosDutResultsDir, tokenFile)
	defer func() {
		// Both errs will be ignored as they are non-critical and logged in the respective calls.
		dutService.Remove(ctx)
		common.AddContentsToLog("log.txt", crosDutResultsDir, "Reading cros-dut log file.")
	}()
	if err != nil {
		return nil, errors.Annotate(err, "run post process, unable to start dut-service").Err()
	}

	log.Printf("Post Process: created the %s directory %s", PostProcessName, postProcessDir)
	request := req.GetRequest()
	if err := writePostProcessInput(inputFileName, request); err != nil {
		return nil, errors.Annotate(err, "prepare to run postprocess: failed to create input file %s", inputFileName).Err()
	}

	if err = services.RunPostProcessCLI(ctx,
		postProcessContainer,
		networkName,
		inputFileName,
		postProcessDir,
		int32(dutService.ServicePort),
		tokenFile); err != nil {
		return nil, errors.Annotate(err, "run postprocess: failed to run %s CLI", PostProcessName).Err()
	}
	resultFileName := path.Join(postProcessDir, "result.json")
	if _, err := os.Stat(resultFileName); os.IsNotExist(err) {
		return nil, errors.Reason("process postprocess result: result not found").Err()
	}
	outp, err := readPostProcessOutput(resultFileName)
	if err != nil {
		return nil, errors.Annotate(err, "process postprocess result: failed to read postprocess output").Err()
	}
	return &api.CrosToolRunnerPostTestResponse{
		Response: outp,
	}, err
}

// writePostProcessInput writes a RunActivitiesRequest json.
func writePostProcessInput(file string, req *api.RunActivitiesRequest) error {
	f, err := os.Create(file)
	if err != nil {
		return errors.Annotate(err, "fail to create file %v", file).Err()
	}
	m := jsonpb.Marshaler{}
	if err := m.Marshal(f, req); err != nil {
		return errors.Annotate(err, "fail to marshal request to file %v", file).Err()
	}
	return nil
}

// readPostProcessOutput reads output file generated by post-process.
func readPostProcessOutput(filePath string) (*api.RunActivitiesResponse, error) {
	r, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Annotate(err, "read output").Err()
	}
	out := &api.RunActivitiesResponse{}
	umrsh := common.JsonPbUnmarshaler()
	err = umrsh.Unmarshal(r, out)
	return out, errors.Annotate(err, "read output").Err()
}
