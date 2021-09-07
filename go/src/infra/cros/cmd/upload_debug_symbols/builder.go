// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package main implements a distributed worker model for uploading debug
// symbols to the crash service. This package will be called by recipes through
// CIPD and will perform the buisiness logic of the builder.
package main

import (
	"context"
	"fmt"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	lgs "go.chromium.org/luci/common/gcloud/gs"
	"infra/cros/internal/gs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	// Default server URLs for the crash service.
	prodUploadUrl    = "https://prod-crashsymbolcollector-pa.googleapis.com/v1"
	stagingUploadUrl = "https://staging-crashsymbolcollector-pa.googleapis.com/v1"
	// Time in milliseconds to sleep before retrying the task.
	sleepTimeMs = 100
)

// taskConfig will contain the information needed to complete the upload task.
type taskConfig struct {
	symbolPath string
	retryCount string
	dryRun     bool
	isStaging  bool
}

// channels will contain the forward configChannel and backwards retryChannel
// that the upload worker will use. The forward channel will have an information
// flow going from the main loop(driver) to the worker. The backwards channel is
// the opposite.
type channels struct {
	configChannel chan taskConfig
	retryChannel  chan taskConfig
}

type uploadDebugSymbols struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	gsPath      string
	workerCount int
	retryCount  int
	isStaging   bool
	dryRun      bool
}

func getCmdUploadDebugSymbols(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "upload <options>",
		ShortDesc: "Upload debug symbols to crash.",
		CommandRun: func() subcommands.CommandRun {
			b := &uploadDebugSymbols{}
			b.authFlags = authcli.Flags{}
			b.authFlags.Register(b.GetFlags(), authOpts)
			b.Flags.StringVar(&b.gsPath, "gs-path", "localhost", ("Url pointing to the GS " +
				"bucket storing the tarball."))
			b.Flags.IntVar(&b.workerCount, "worker-count", -1, ("Number of worker threads" +
				" to spawn."))
			b.Flags.IntVar(&b.retryCount, "retry-count", -1, ("Number of total upload retries" +
				" allowed."))
			b.Flags.BoolVar(&b.isStaging, "is-staging", false, ("Specifies if the builder" +
				" should push to the staging crash service or prod."))
			b.Flags.BoolVar(&b.dryRun, "dry-run", false, ("Specified whether network" +
				" operations should be dry ran."))
			return b
		}}
}

// generateClient handles the authentication of the user then generation of the
// client to be used by the gs module.
func generateClient(ctx context.Context, authOpts auth.Options) (*gs.ProdClient, error) {
	authedClient, err := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts).Client()
	if err != nil {
		return nil, err
	}

	gsClient, err := gs.NewProdClient(ctx, authedClient)
	if err != nil {
		return nil, err
	}
	return gsClient, err
}

// fetchTgz will download the tarball from google storage which contains all
// of the symbol files to be uploaded. Once downloaded it will return the local
// filepath to tarball.
func fetchTgz(client gs.Client, gsPath, tgzPath string) error {
	// TODO(b/197010274): remove skeleton code.
	return client.Download(lgs.Path(gsPath), tgzPath)
}

// uploadWorker will perform the upload of the symbol file to the crash service.
func uploadWorker(chans channels) error {
	// Fetch the local file from the unpacked tarball.

	// Open up an https request to the crash service.

	// Verify if the file has been uploaded already.

	// Upload the file.

	// Return with appropriate status code.
	// TODO(b/197010274): remove skeleton code.
	return nil
}

// unpackTarball will take the local path of the fetched tarball and then unpack
// it. It will then return a list of file paths pointing to the unpacked symbol
// files.
func unzipTgz(inputPath, outputPath string) error {
	// TODO(b/197010274): remove skeleton code.
	return nil
}

// unpackTarball will take the local path of the fetched tarball and then unpack
// it. It will then return a list of file paths pointing to the unpacked symbol
// files.
func unpackTarball(inputPath, outputDir string) ([]string, error) {
	// TODO(b/197010274): remove skeleton code.
	return []string{"./path"}, nil
}

// generateConfigs will take a list of strings with containing the paths to the
// unpacked symbol files. It will return a list of generated task configs
// alongside the communication channels to be used.
func generateConfigs(symbolFiles []string) ([]taskConfig, *channels, error) {
	// TODO(b/197010274): remove skeleton code.
	return nil, nil, nil
}

// doUpload is the main loop that will spawn goroutines that will handle the
// upload tasks. Should the worker fail it's upload and we have retries left,
// send the task to the end of the channel's buffer.
func doUpload(tasks []taskConfig, chans *channels, retryCount int,
	isStaging, dryRun bool) (int, error) {
	// TODO(b/197010274): remove skeleton code.
	return 0, nil
}

// validate checks the values of the required flags and returns an error they
// aren't populated. Since multiple flags are required, the error message may
// include multiple error statements.
func (b *uploadDebugSymbols) validate() error {
	errStr := ""
	if b.gsPath == "localhost" {
		errStr = fmt.Sprintf("error: --gs-path value is required.\n")
	}
	if b.workerCount == -1 {
		errStr = fmt.Sprintf(errStr, "error: --worker-count value is required.\n")
	}
	if b.retryCount == -1 {
		errStr = fmt.Sprintf(errStr, "error: --retry-count value is required.\n")
	}

	if errStr != "" {
		return fmt.Errorf(errStr)
	}
	return nil
}

// main is the function to be called by the CLI execution.
func (b *uploadDebugSymbols) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	// Generate authenticated http client.
	ctx := context.Background()
	authOpts, err := b.authFlags.Options()
	if err != nil {
		log.Fatal(err)
	}
	client, err := generateClient(ctx, authOpts)
	if err != nil {
		log.Fatal(err)
	}
	// Create local dir and file for tarball to live in.
	workDir, err := ioutil.TempDir("", "tarball")
	if err != nil {
		log.Fatal(err)
	}

	tgzPath := filepath.Join(workDir, "debug.tgz")
	tarbalPath := filepath.Join(workDir, "debug.tar")
	fmt.Print(tgzPath + "\n")
	defer os.RemoveAll(workDir)

	err = fetchTgz(client, b.gsPath, tgzPath)
	if err != nil {
		log.Fatal(err)
	}

	err = unzipTgz(tgzPath, tarbalPath)
	symbolFiles, err := unpackTarball(tarbalPath, workDir)

	if err != nil {
		log.Fatal(err)
	}

	tasks, chans, err := generateConfigs(symbolFiles)

	retcode, err := doUpload(tasks, chans, b.retryCount, b.isStaging, b.dryRun)

	if err != nil {
		log.Fatal(err)
	}
	// TODO(b/197010274): remove skeleton code.
	// Return:
	// 		0: Success, all symbols uploaded.
	// 		1: Failure, more failures occurred than retries were allotted
	return retcode
}
