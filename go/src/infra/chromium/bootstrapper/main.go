// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	stderrors "errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"infra/chromium/bootstrapper/bootstrap"
	"infra/chromium/bootstrapper/clients/gclient"
	"infra/chromium/bootstrapper/clients/gerrit"
	"infra/chromium/bootstrapper/clients/gitiles"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	logdogbootstrap "go.chromium.org/luci/logdog/client/butlerlib/bootstrap"
	"go.chromium.org/luci/logdog/client/butlerlib/streamclient"
	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type getOptionsFn func() options

func parseFlags() options {
	outputPath := flag.String("output", "", "Path to write the final build.proto state to.")
	polymorphic := flag.Bool("polymorphic", false, "Whether the builder bootstraps properties for other builders instead of itself; polymorphic builders give precedence to build properties rather than the properties in the properties file")
	propertiesOptional := flag.Bool("properties-optional", false, "Whether missing $bootstrap/properties property should be allowed")
	flag.Parse()
	return options{
		outputPath:         *outputPath,
		packagesRoot:       "packages",
		polymorphic:        *polymorphic,
		propertiesOptional: *propertiesOptional,
	}
}

func getBuild(ctx context.Context, input io.Reader) (*buildbucketpb.Build, error) {
	logging.Infof(ctx, "reading build input")
	data, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, errors.Annotate(err, "failed to read build input").Err()
	}
	logging.Infof(ctx, "unmarshalling build input")
	build := &buildbucketpb.Build{}
	if err = proto.Unmarshal(data, build); err != nil {
		return nil, errors.Annotate(err, "failed to unmarshall build").Err()
	}
	return build, nil
}

type options struct {
	outputPath         string
	packagesRoot       string
	polymorphic        bool
	propertiesOptional bool
}

type bootstrapFn func(ctx context.Context, input io.Reader, opts options) ([]string, []byte, error)

func performBootstrap(ctx context.Context, input io.Reader, opts options) ([]string, []byte, error) {
	build, err := getBuild(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	logging.Infof(ctx, "creating bootstrap input")
	inputOpts := bootstrap.InputOptions{
		Polymorphic:        opts.polymorphic,
		PropertiesOptional: opts.propertiesOptional,
	}
	bootstrapInput, err := inputOpts.NewInput(build)
	if err != nil {
		return nil, nil, err
	}

	var config *bootstrap.BootstrapConfig

	var exe *bootstrap.BootstrappedExe
	var cmd []string

	// Downloading the necessary packages and getting the appropriate properties both speak to
	// external services but don't necessarily depend on each other, so use an errgroup to do
	// them in parallel

	// Introduce a new block to shadow the ctx variable so that the outer
	// value can't be used accidentally
	{
		group, ctx := errgroup.WithContext(ctx)

		// If the builder's properties are in a dependent project, getting the properties
		// might require the gclient binary from the depot_tools package, so provide a
		// channel that can be used to synchronize where necessary
		depotToolsCh := make(chan string, 1)

		group.Go(func() error {
			logging.Infof(ctx, "downloading necessary packages")
			var err error
			exe, cmd, err = bootstrap.DownloadPackages(ctx, bootstrapInput, opts.packagesRoot, map[string]chan<- string{
				bootstrap.DepotToolsId: depotToolsCh,
			})
			return errors.Annotate(err, "failed to download necessary packages").Err()
		})

		group.Go(func() error {
			// gclientGetter will only be called if dependency_project is set in the
			// $bootstrap/exe property, depot_tools will always be downloaded in that
			// case
			gclientGetter := func(ctx context.Context) (*gclient.Client, error) {
				var depotToolsPackagePath string
				select {
				case depotToolsPackagePath = <-depotToolsCh:
				case <-ctx.Done():
					return nil, ctx.Err()
				}
				gclientPath := filepath.Join(depotToolsPackagePath, "depot_tools", "gclient")
				return gclient.NewClient(gclientPath), nil
			}
			bootstrapper := bootstrap.NewBuildBootstrapper(gitiles.NewClient(ctx), gerrit.NewClient(ctx), gclientGetter)

			logging.Infof(ctx, "getting bootstrapped config")
			var err error
			config, err = bootstrapper.GetBootstrapConfig(ctx, bootstrapInput)
			return err
		})

		if err := group.Wait(); err != nil {
			return nil, nil, err
		}
	}

	logging.Infof(ctx, "updating build")
	err = config.UpdateBuild(build, exe)
	if err != nil {
		return nil, nil, err
	}

	logging.Infof(ctx, "marshalling bootstrapped build input")
	recipeInput, err := proto.Marshal(build)
	if err != nil {
		return nil, nil, errors.Annotate(err, "failed to marshall bootstrapped build input: <%s>", build).Err()
	}

	if opts.outputPath != "" {
		cmd = append(cmd, "--output", opts.outputPath)
	}

	return cmd, recipeInput, nil
}

type executeCmdFn func(ctx context.Context, cmd []string, input []byte) error

func executeCmd(ctx context.Context, cmd []string, input []byte) error {
	cmdCtx := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	cmdCtx.Stdin = bytes.NewBuffer(input)
	cmdCtx.Stdout = os.Stdout
	cmdCtx.Stderr = os.Stderr
	return cmdCtx.Run()
}

type updateBuildFn func(ctx context.Context, build *buildbucketpb.Build) error

func updateBuild(ctx context.Context, build *buildbucketpb.Build) (err error) {
	outputData, err := proto.Marshal(build)
	if err != nil {
		return errors.Annotate(err, "failed to marshal output build.proto").Err()
	}

	logdog, err := logdogbootstrap.Get()
	if err != nil {
		return errors.Annotate(err, "failed to get logdog bootstrap instance").Err()
	}
	stream, err := logdog.Client.NewDatagramStream(
		ctx,
		luciexe.BuildProtoStreamSuffix,
		streamclient.WithContentType(luciexe.BuildProtoContentType),
	)
	if err != nil {
		return errors.Annotate(err, "failed to get datagram stream").Err()
	}
	defer func() {
		closeErr := stream.Close()
		if closeErr != nil {
			if err != nil {
				logging.Errorf(ctx, closeErr.Error())
			} else {
				err = closeErr
			}
		}
	}()

	err = stream.WriteDatagram(outputData)
	if err != nil {
		err = errors.Annotate(err, "failed to write modified build").Err()
		return
	}
	return
}

func bootstrapMain(ctx context.Context, getOpts getOptionsFn, performBootstrap bootstrapFn, executeCmd executeCmdFn, updateBuild updateBuildFn) error {
	opts := getOpts()
	cmd, input, err := performBootstrap(ctx, os.Stdin, opts)
	if err == nil {
		logging.Infof(ctx, "executing %s", cmd)
		err = executeCmd(ctx, cmd, input)
	}

	if err != nil {
		logging.Errorf(ctx, err.Error())
		// An ExitError indicates that we were able to bootstrap the
		// executable and that it failed, as opposed to being unable to
		// launch the bootstrapped executable
		var exitErr *exec.ExitError
		if !stderrors.As(err, &exitErr) {
			build := &buildbucketpb.Build{}
			build.SummaryMarkdown = fmt.Sprintf("<pre>%s</pre>", err)
			build.Status = buildbucketpb.Status_INFRA_FAILURE
			if bootstrap.PatchRejected.In(err) {
				build.Output = &buildbucketpb.Build_Output{
					Properties: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"failure_type": structpb.NewStringValue("PATCH_FAILURE"),
						},
					},
				}
			}
			if err := updateBuild(ctx, build); err != nil {
				logging.Errorf(ctx, errors.Annotate(err, "failed to update build with failure details").Err().Error())
			}
		}
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()
	ctx = gologger.StdConfig.Use(ctx)

	// Tracking soft deadline and calling shutdown causes the bootstrapper
	// to participate in the termination protocol. No explicit action is
	// necessary to terminate the bootstrapped executable, the signal will
	// be propagated to the entire process/console group.
	ctx, shutdown := lucictx.TrackSoftDeadline(ctx, 500*time.Millisecond)
	defer shutdown()

	if err := bootstrapMain(ctx, parseFlags, performBootstrap, executeCmd, updateBuild); err != nil {
		os.Exit(1)
	}
}
