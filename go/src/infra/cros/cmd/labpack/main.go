// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

/*
The labpack program allows to run repair tasks for ChromeOS devices in the lab.
For more information please read go/paris-.
Managed by Chrome Fleet Software (go/chrome-fleet-software).
*/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	luciauth "go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/errors"
	lucigs "go.chromium.org/luci/common/gcloud/gs"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
	"google.golang.org/protobuf/encoding/protojson"

	"infra/cros/cmd/labpack/internal/site"
	steps "infra/cros/cmd/labpack/internal/steps"
	"infra/cros/cmd/labpack/internal/tlw"
	kclient "infra/cros/karte/client"
	"infra/cros/recovery"
	"infra/cros/recovery/karte"
	"infra/cros/recovery/logger"
	"infra/cros/recovery/logger/metrics"
	"infra/cros/recovery/upload"
	"infra/libs/skylab/buildbucket"
	ufsUtil "infra/unifiedfleet/app/util"
)

// DescribeMyDirectoryAndEnvironment controls whether labpack should write information
// about where it was run (cwd), what files are near it, and the contents of the environment.
const DescribeMyDirectoryAndEnvironment = true

// DescriptionCommand describes the environment where labpack was run. It must write all of its output to stdout.
const DescriptionCommand = `( echo BEGIN; echo PWD; pwd ; echo FIND; find . ; echo ENV; env; echo END )`

type ResponseUpdater func(*steps.LabpackResponse)

func main() {
	log.SetPrefix(fmt.Sprintf("%s: ", filepath.Base(os.Args[0])))
	log.Printf("Running version: %s", site.VersionNumber)
	log.Printf("Running in buildbucket mode")

	input := &steps.LabpackInput{}
	var writeOutputProps ResponseUpdater
	var mergeOutputProps ResponseUpdater
	build.Main(input, &writeOutputProps, &mergeOutputProps,
		func(ctx context.Context, args []string, state *build.State) error {
			// Right after instantiating the logger, but inside build.Main's callback,
			// make sure that we log what our environment looks like.
			if DescribeMyDirectoryAndEnvironment {
				describeEnvironment(os.Stderr)
				// Describe the contents of the directory once on the way out too.
				// We will use this information to decide what to persist.
				defer describeEnvironment(os.Stderr)
			}

			// Set the log (via the Go standard library's log package) to Stderr, since we know that stderr is collected
			// for the process as a whole.
			log.SetOutput(os.Stderr)

			// We need more logging in order to fix some data gaps (like the lack of a buildbucket ID).
			// Log the input string as JSON so we can see exactly which fields are populated with what in prod.
			b := protojson.MarshalOptions{
				Indent: "  ",
			}.Format(input)
			log.Printf("%s\n", string(b))

			err := mainRunInternal(ctx, input, state, writeOutputProps)
			return errors.Annotate(err, "main").Err()
		},
	)
	log.Printf("Labpack done!")
}

// mainRun runs function for BB and provide result.
func mainRunInternal(ctx context.Context, input *steps.LabpackInput, state *build.State, writeOutputProps ResponseUpdater) error {
	// Result errors which specify the result of main run.
	var resultErrors []error

	logRoot, err := getTaskDir()
	if err != nil {
		return errors.Annotate(err, "main run internal").Err()
	}
	// TODO: Change level to Info when all logging files will be upload to GC.
	ctx, lg, err := createLogger(ctx, logRoot, logging.Debug)
	if err != nil {
		return errors.Annotate(err, "main run internal").Err()
	}
	defer func() { lg.Close() }()

	// Run recovery lib and get response.
	// Set result as fail by default in case it fail to finish by some reason.
	res := &steps.LabpackResponse{
		Success:    false,
		FailReason: "Fail by unknown reason!",
	}
	defer func() {
		// Write result as last step.
		writeOutputProps(res)
	}()
	lg.Infof("Prepare print input params...")
	if err = printInputs(ctx, input); err != nil {
		lg.Debugf("main run internal: failed to marshal proto. Error: %s", err)
		return err
	}
	var metrics metrics.Metrics
	if !input.GetNoMetrics() {
		lg.Infof("Prepare create Karte client...")
		var err error
		metrics, err = karte.NewMetrics(ctx, kclient.ProdConfig(luciauth.Options{}))
		if err == nil {
			lg.Infof("Karte client successfully created.")
		} else {
			lg.Errorf("Failed to instantiate Karte client: %s", err)
			resultErrors = append(resultErrors, err)
		}
	}
	lg.Infof("Starting task execution...")
	if err := internalRun(ctx, input, state, metrics, lg, logRoot); err != nil {
		res.Success = false
		res.FailReason = err.Error()
		resultErrors = append(resultErrors, err)
	}
	lg.Infof("Finished task execution.")
	lg.Infof("Starting uploading logs...")
	if err := uploadLogs(ctx, state, lg); err != nil {
		res.Success = false
		if len(resultErrors) == 0 {
			// We should not override runerror reason as it more important.
			// If upload logs error is only exits then set it as reason.
			res.FailReason = err.Error()
		}
		resultErrors = append(resultErrors, err)
	}
	lg.Infof("Finished uploading logs.")
	// if err is nil then will marked as SUCCESS
	if len(resultErrors) == 0 {
		// Reset reason and state as no errors detected.
		res.Success = true
		res.FailReason = ""
		return nil
	}
	return errors.Annotate(errors.MultiError(resultErrors), "run recovery").Err()
}

// Upload logs to google cloud.
func uploadLogs(ctx context.Context, state *build.State, lg logger.Logger) (rErr error) {
	step, ctx := build.StartStep(ctx, "Upload logs")
	lg.Infof("Beginning to upload logs")
	defer func() {
		if r := recover(); r != nil {
			lg.Debugf("Received panic: %v\n", r)
			rErr = errors.Reason("panic: %v", r).Err()
		}
		lg.Infof("Finished uploading logs: ok=%t.", rErr == nil)
		step.End(rErr)
	}()
	// Construct the client that we will need to push the logs first.
	authenticator := luciauth.NewAuthenticator(
		ctx,
		luciauth.SilentLogin,
		luciauth.Options{
			Scopes: []string{
				luciauth.OAuthScopeEmail,
				"https://www.googleapis.com/auth/devstorage.read_write",
			},
		},
	)
	if authenticator != nil {
		lg.Infof("NewAuthenticator(...): successfully authed!")
	} else {
		return errors.Reason("NewAuthenticator(...): did not successfully auth!").Err()
	}
	email, err := authenticator.GetEmail()
	if err != nil {
		return errors.Annotate(err, "upload logs").Err()
	}
	lg.Infof("Auth email is %q", email)

	rt, err := authenticator.Transport()
	if err != nil {
		return errors.Annotate(err, "authenticator.Transport(...): error").Err()
	}

	client, err := lucigs.NewProdClient(ctx, rt)
	if err != nil {
		return errors.Annotate(err, "failed to create client(...)").Err()
	}

	lg.Infof("Persist the swarming logs")
	// Actually persist the logs.
	swarmingTaskID := state.Build().GetInfra().GetSwarming().GetTaskId()
	gsURL, err := parallelUpload(ctx, lg, client, swarmingTaskID)
	if err != nil {
		return errors.Annotate(err, "upload logs").Err()
	}
	// Set the summary markdown to something noticeable.
	// In the future, change this to be a link to the logs.
	step.Modify(func(sv *build.StepView) {
		u := strings.TrimPrefix(gsURL, "gs://")
		u = fmt.Sprintf("https://%s/%s", "stainless.corp.google.com/browse", u)
		sv.SummaryMarkdown = fmt.Sprintf("[GS logs](%s)", u)
	})
	return nil
}

// parallelUpload performs an upload in parallel to the google-storage bucket.
//
// parallelUpload will fail when given invalid arguments. However, it will not fail
// simply because the upload attempt was unsuccessful.
func parallelUpload(ctx context.Context, lg logger.Logger, client lucigs.Client, swarmingTaskID string) (string, error) {
	if lg == nil {
		return "", errors.Reason("parallel-upload: logger cannot be nil").Err()
	}
	if client == nil {
		return "", errors.Reason("paralel-upload: client cannot be nil").Err()
	}
	if swarmingTaskID == "" {
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		lg.Errorf("Swarming task is empty. Falling back to timestamp %q.", timestamp)
		swarmingTaskID = fmt.Sprintf("FAKE-ID-%s", timestamp)
	}
	// upload.Upload can potentially run for a long time. Set a timeout of 30s.
	//
	// upload.Upload does respond to cancellation (which callFuncWithTimeout uses internally), but
	// the correct of this code does not and should not depend on this fact.
	//
	// callFuncWithTimeout synchronously calls a function with a timeout and then unconditionally hands control
	// back to its caller. The goroutine that's created in the background will not by itself keep the process alive.
	// TODO(gregorynisbet): Allow this parameter to be overridden from outside.
	// TODO(crbug/1311842): Switch this bucket back to chromeos-autotest-results.
	gsURL := fmt.Sprintf("gs://chrome-fleet-karte-autotest-results/swarming-%s", swarmingTaskID)
	lg.Infof("Swarming task %q is non-empty. Uploading to %q", swarmingTaskID, gsURL)
	status, err := callFuncWithTimeout(ctx, 5*time.Minute, func(ctx context.Context) error {
		lg.Infof("Beginning upload attempt. Starting five minute timeout.")
		lg.Infof("Writing upload marker.")
		// TODO(b:227489086): Remove this file.
		if wErr := os.WriteFile("_labpack_upload_marker", []byte("ca85a1f7-0de3-43c5-90ff-2e00b1041007"), 0o666); wErr != nil {
			lg.Errorf("Failed to write upload marker file: %s", wErr)
		}

		lg.Infof("Calling upload.")
		return upload.Upload(ctx, client, &upload.Params{
			// TODO(gregorynisbet): Change this to the log root.
			SourceDir:         ".",
			GSURL:             gsURL,
			MaxConcurrentJobs: 10,
		})
	})
	lg.Infof("Upload log subtask status: %s", status)
	if err != nil {
		// TODO: Register error to Karte.
		lg.Errorf("Upload task error: %s", err)
	}
	return gsURL, nil
}

// internalRun main entry point to execution received request.
func internalRun(ctx context.Context, in *steps.LabpackInput, state *build.State, metrics metrics.Metrics, lg logger.Logger, logRoot string) (err error) {
	defer func() {
		// Catching the panic here as luciexe just set a step as fail and but not exit execution.
		lg.Debugf("Checking if there is a panic!")
		if r := recover(); r != nil {
			lg.Debugf("Received panic: %v\n", r)
			err = errors.Reason("panic: %v", r).Err()
		}
	}()
	ctx = setupContextNamespace(ctx, ufsUtil.OSNamespace)
	ctx, access, err := tlw.NewAccess(ctx, in)
	if err != nil {
		return errors.Annotate(err, "internal run").Err()
	}
	defer func() { access.Close(ctx) }()

	// Recovery is the task that we want 90% of the time. However, silently making
	// recovery the default can cause us to silently fall back to performing a recovery task
	// when we did not intend to, which is hard to discover unless you carefully read the logs.
	//
	// To avoid this, I am making the logic here much stricter and ending the task early if
	// we use an unrecognized (or empty) task name.
	task, ok := supportedTasks[in.TaskName]
	if !ok {
		return errors.Reason("task name %q is invalid", in.TaskName).Err()
	}
	infraPb := state.Build().GetInfra()

	runArgs := &recovery.RunArgs{
		UnitName:              in.GetUnitName(),
		TaskName:              task,
		Access:                access,
		Logger:                lg,
		ShowSteps:             !in.GetNoStepper(),
		Metrics:               metrics,
		EnableRecovery:        in.GetEnableRecovery(),
		EnableUpdateInventory: in.GetUpdateInventory(),
		SwarmingTaskID:        infraPb.GetSwarming().GetTaskId(),
		BuildbucketID:         infraPb.GetBackend().GetTask().GetId().GetId(),
		LogRoot:               logRoot,
	}
	if uErr := runArgs.UseConfigBase64(in.GetConfiguration()); uErr != nil {
		return uErr
	}

	lg.Debugf("Labpack: starting the task...")
	if err := recovery.Run(ctx, runArgs); err != nil {
		lg.Debugf("Labpack: finished task run with error: %v", err)
		return errors.Annotate(err, "internal run").Err()
	}
	lg.Debugf("Labpack: finished task successful!")
	return nil
}

// Mapping of all supported tasks.
var supportedTasks = map[string]buildbucket.TaskName{
	string(buildbucket.AuditRPM):     buildbucket.AuditRPM,
	string(buildbucket.AuditStorage): buildbucket.AuditStorage,
	string(buildbucket.AuditUSB):     buildbucket.AuditUSB,
	string(buildbucket.Custom):       buildbucket.Custom,
	string(buildbucket.Deploy):       buildbucket.Deploy,
	string(buildbucket.Recovery):     buildbucket.Recovery,
	string(buildbucket.DeepRecovery): buildbucket.DeepRecovery,
}
