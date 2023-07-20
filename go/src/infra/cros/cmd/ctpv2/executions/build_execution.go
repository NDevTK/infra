// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executions

import (
	"context"
	"fmt"
	"log"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	"go.chromium.org/luci/common/errors"
	"infra/cros/cmd/common_lib/common"
	"infra/cros/cmd/common_lib/common_configs"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/protos"
	"infra/cros/cmd/ctpv2/internal/configs"
	"infra/cros/cmd/ctpv2/internal/data"
)

// TODO : Re-structure different execution flow properly later.
// LuciBuildExecution represents build executions.
func LuciBuildExecution() {
	// Set input property reader functions
	var ctrCipdInfoReader func(context.Context) *protos.CipdVersionInfo
	build.MakePropertyReader(common.HwTestCtrInputPropertyName, &ctrCipdInfoReader)
	input := &api.CTPv2Request{}

	// Set output props writer functions
	// TODO: add the fields to the response that is responsible for
	// feeding the test results to upstream.
	var writeOutputProps func(*api.CTPv2Response)
	var mergeOutputProps func(*api.CTPv2Response)

	build.Main(input, &writeOutputProps, &mergeOutputProps,
		func(ctx context.Context, args []string, st *build.State) error {
			log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmsgprefix)
			logging.Infof(ctx, "have input %v", input)
			ctrCipdInfo := ctrCipdInfoReader(ctx)
			logging.Infof(ctx, "have ctr info: %v", ctrCipdInfo)
			logging.Infof(ctx, "ctr label: %s", ctrCipdInfo.GetVersion().GetCipdLabel())
			resp := &api.CTPv2Response{}
			resp, err := executeFiltersInLuciBuild(ctx, input, ctrCipdInfo.GetVersion().GetCipdLabel(), st)
			// TODO: add compressed result
			// if resp != nil {
			// 	m, _ := proto.Marshal(resp)
			// 	var b bytes.Buffer
			// 	w := zlib.NewWriter(&b)
			// 	_, _ = w.Write(m)
			// 	_ = w.Close()
			//
			// 	resp.CompressedResult = base64.StdEncoding.EncodeToString(b.Bytes())
			// }
			if err != nil {
				logging.Infof(ctx, "error found: %s", err)
				st.SetSummaryMarkdown(err.Error())
				// resp.ErrorSummaryMarkdown = err.Error()
			}

			writeOutputProps(resp)
			return err
		},
	)
}

func executeFiltersInLuciBuild(
	ctx context.Context,
	req *api.CTPv2Request,
	ctrCipdVersion string,
	buildState *build.State) (*api.CTPv2Response, error) {
	// Validation
	if ctrCipdVersion == "" {
		return nil, fmt.Errorf("Cros-tool-runner cipd version cannot be empty for hw test execution.")
	}

	// Create ctr
	ctrCipdInfo := crostoolrunner.CtrCipdInfo{
		Version:        ctrCipdVersion,
		CtrCipdPackage: common.CtrCipdPackage,
	}

	ctr := &crostoolrunner.CrosToolRunner{
		CtrCipdInfo:       ctrCipdInfo,
		EnvVarsToPreserve: []string{},
	}

	// Fetch container MD

	// TODO: currently there is a bit of a race between getting the metadata to run test-finder, and the boards provided
	// to the request. We don't want to run test-finder per board as that's extremely expensive to setup/act on, however
	// CFT design has test-finder being board specific. For initial MVP we will just use the first board in the request to
	// get the container MD from, but this will need to be solved long term.
	board, gcsPath, err := gcsInfo(req)
	if err != nil {
		return nil, err
	}
	containerMetadata, err := common.FetchImageData(ctx, board, gcsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch image data: %s", err)
	}

	dockerKeyFile, err := common.LocateFile([]string{common.LabDockerKeyFileLocation, common.VmLabDockerKeyFileLocation})
	if err != nil {
		return nil, fmt.Errorf("unable to locate dockerKeyFile during initialization: %w", err)
	}

	containerCfg := common_configs.NewContainerConfig(ctr, containerMetadata, false)
	executorCfg := configs.NewExecutorConfig(ctr, containerCfg)
	cmdCfg := configs.NewCommandConfig(executorCfg)

	// Create state keeper
	sk := &data.FilterStateKeeper{
		CtpV2Req:              req,
		DockerKeyFileLocation: dockerKeyFile,
		Ctr:                   ctr,
	}

	// Generate config
	ctpv2Config := configs.NewCtpv2ExecutionConfig(configs.LuciBuildFilterExecutionConfigType, cmdCfg, sk)
	err = ctpv2Config.GenerateConfig(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "error during generating filter configs: ").Err()
	}

	// Execute config
	err = ctpv2Config.Execute(ctx)
	if err != nil {
		return sk.CtpV2Response, errors.Annotate(err, "error during executing hw test configs: ").Err()
	}

	return sk.CtpV2Response, nil
}

func gcsInfo(req *api.CTPv2Request) (string, string, error) {
	board := getFirstBoardFromLegacy(req.Targets)
	if board == "" {
		return "", "", errors.New("no board provided in legacy request")
	}

	gcsPath := getFirstGcsPathFromLegacy(req.Targets)
	if gcsPath == "" {
		return "", "", errors.New("no gcsPath provided in legacy request")
	}

	return board, gcsPath, nil
}

func getFirstBoardFromLegacy(targs []*api.Targets) string {
	if len(targs) == 0 {
		return ""
	}
	switch hw := targs[0].HwTarget.Target.(type) {
	case *api.HWTarget_LegacyHw:
		return hw.LegacyHw.Board
	default:
		return ""
	}
}

func getFirstGcsPathFromLegacy(targs []*api.Targets) string {
	if len(targs) == 0 {
		return ""
	}
	if len(targs[0].SwTargets) == 0 {
		return ""
	}
	switch sw := targs[0].SwTargets[0].SwTarget.(type) {
	case *api.SWTarget_LegacySw:
		return sw.LegacySw.GcsPath
	default:
		return ""
	}
}
