// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/system/environ"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe"

	"infra/experimental/golangbuild/golangbuildpb"
)

// gomoteSetup sets up the environment for a gomote then invokes the command
// in args. This path must closely, if not identically, match the setup path
// before calling into one of the mode-specific runners.
func gomoteSetup(ctx context.Context, builderName string, args []string) error {
	// Define working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return infraErrorf("get CWD")
	}

	log.Printf("setting up build environment for gomote at %s...", cwd)

	// Set up the basic parts of the environment first.
	tmpDir := filepath.Join(cwd, "tmp")
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return err
	}
	cacheDir := filepath.Join(cwd, "cache")
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return err
	}

	// Set all luciexe temporary directories as well.
	for _, key := range luciexe.TempDirEnvVars {
		ctx = addEnv(ctx, fmt.Sprintf("%s=%s", key, tmpDir))
	}

	// Set up LUCI_CONTEXT.
	ctx = lucictx.SetLUCIExe(ctx, &lucictx.LUCIExe{
		CacheDir: cacheDir,
	})

	log.Printf("obtaining builder info for %s...", builderName)

	// Get builder info. In this context, we have to contact buildbucket,
	// since we're assuming that we're not participating in the luciexe
	// protocol.
	inputs, experiments, err := getBuilderInfo(ctx, builderName)
	if err != nil {
		return infraErrorf("obtaining builder info: %w", err)
	}

	log.Printf("installing tools...")

	// Install some tools we'll need, including a bootstrap toolchain.
	toolsRoot, err := installTools(ctx, inputs, experiments)
	if err != nil {
		return infraErrorf("installing tools: %w", err)
	}

	// Install tools in context.
	ctx = withToolsRoot(ctx, toolsRoot)

	// Get the CAS instance.
	ctx, err = casInstanceFromEnv(ctx)
	if err != nil {
		return infraErrorf("casInstanceFromEnv: %w", err)
	}

	// Set up the rest of the environment.
	goroot := filepath.Join(cwd, "goroot")
	gopath := filepath.Join(cwd, "gopath")
	gocacheDir := filepath.Join(cwd, "gocache")
	ctx = setupEnv(ctx, inputs, builderName, goroot, gopath, gocacheDir)

	// Log the environment changes.
	want := environ.FromCtx(ctx)
	base := environ.System()
	log.Printf("environment changes:\n%s", diffEnv(base, want))

	// Execute the command in args.
	cmd := command(ctx, args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("invoking %s...", cmd.String())
	return cmd.Run()
}

func getBuilderInfo(ctx context.Context, builderName string) (*golangbuildpb.Inputs, map[string]struct{}, error) {
	id, err := parseBuilderID(builderName)
	if err != nil {
		return nil, nil, err
	}

	// Contact buildbucket to obtain information about the builder.
	host := chromeinfra.BuildbucketHost
	if bbCtx := lucictx.GetBuildbucket(ctx); bbCtx != nil {
		if bbCtx.GetHostname() != "" {
			host = bbCtx.Hostname
		}
	}
	hc, err := createAuthenticator(ctx).Client()
	if err != nil {
		return nil, nil, fmt.Errorf("authenticator.Client: %w", err)
	}
	bc := bbpb.NewBuildersPRPCClient(&prpc.Client{
		C:       hc,
		Host:    host,
		Options: prpc.DefaultOptions(),
	})
	b, err := bc.GetBuilder(ctx, &bbpb.GetBuilderRequest{Id: id})
	if err != nil {
		return nil, nil, fmt.Errorf("getting builder info for %q: %w", builderName, err)
	}

	// Parse the properties out of the JSON string in the builder config.
	inputs := new(golangbuildpb.Inputs)
	if err := json.Unmarshal([]byte(b.GetConfig().GetProperties()), inputs); err != nil {
		return nil, nil, err
	}

	// Collect all experiments, but only pass through those that are on 100% of the time.
	experiments := make(map[string]struct{})
	for name, chance := range b.GetConfig().GetExperiments() {
		if chance != 100 {
			continue
		}
		experiments[name] = struct{}{}
	}
	return inputs, experiments, nil
}

func parseBuilderID(builderName string) (*bbpb.BuilderID, error) {
	c := strings.SplitN(builderName, "/", 3)
	if len(c) != 3 {
		return nil, fmt.Errorf("builder name must consist of 3 sections: <project>/<bucket>/<name>")
	}
	return &bbpb.BuilderID{
		Project: c[0],
		Bucket:  c[1],
		Builder: c[2],
	}, nil
}

func diffEnv(base, want environ.Env) string {
	// Find the likely verb for exporting environment variables.
	verb := "export"
	if runtime.GOOS == "windows" {
		verb = "set"
	}

	var sb strings.Builder
	// First, emit all new environment variables or modified ones.
	_ = want.Iter(func(name, value string) error {
		if baseValue, ok := base.Lookup(name); !ok || value != baseValue {
			fmt.Fprintf(&sb, "%s %s=%s\n", verb, name, value)
		}
		return nil
	})
	// Next, explicitly unset environment variables that are not present in want.
	_ = base.Iter(func(name, baseValue string) error {
		if _, ok := want.Lookup(name); !ok {
			fmt.Fprintf(&sb, "%s %s=\n", verb, name)
		}
		return nil
	})
	return sb.String()
}
