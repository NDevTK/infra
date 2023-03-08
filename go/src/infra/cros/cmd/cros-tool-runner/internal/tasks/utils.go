// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"infra/cros/cmd/cros-tool-runner/internal/common"

	build_api "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/tsmon"
	"go.chromium.org/luci/common/tsmon/target"
	"go.chromium.org/luci/lucictx"
)

// readContainersMetadata reads the jsonproto at path containers metadata file.
func readContainersMetadata(p string) (*build_api.ContainerMetadata, error) {
	in := &build_api.ContainerMetadata{}
	r, err := os.Open(p)
	if err != nil {
		return nil, errors.Annotate(err, "read container metadata %q", p).Err()
	}

	umrsh := common.JsonPbUnmarshaler()
	err = umrsh.Unmarshal(r, in)
	return in, errors.Annotate(err, "read container metadata %q", p).Err()
}

func findContainer(cm *build_api.ContainerMetadata, lookupKey, name string) (*build_api.ContainerImageInfo, error) {
	containers := cm.GetContainers()
	if containers == nil {
		return nil, nil
	}
	imageMap, ok := containers[lookupKey]
	if !ok {
		log.Printf("Image %q not found", name)
		return nil, fmt.Errorf("Image %q not found for lookupkey %s", name, lookupKey)
	}
	return imageMap.Images[name], nil
}

func useSystemAuth(ctx context.Context, authFlags *authcli.Flags) (context.Context, error) {
	authOpts, err := authFlags.Options()
	if err != nil {
		return nil, errors.Annotate(err, "switching to system auth").Err()
	}

	authCtx, err := lucictx.SwitchLocalAccount(ctx, "system")
	if err == nil {
		// If there's a system account use it (the case of running on Swarming).
		// Otherwise default to user credentials (the local development case).
		authOpts.Method = auth.LUCIContextMethod
		return authCtx, nil
	}
	log.Printf("System account not found, err %s.\nFalling back to user credentials for auth.\n", err)
	return ctx, nil
}

func metricsInit(ctx context.Context) error {
	// Application level flags.
	log.Printf("Setting up CTR docker tsmon...")

	credsFile := "/creds/service_accounts/service_account_prodx_mon.json"

	if _, err := os.Stat(credsFile); err != nil {
		return errors.Annotate(err, "failed to find tsmon creds file %s", credsFile).Err()
	}

	tsmonFlags := tsmon.NewFlags()
	tsmonFlags.Endpoint = "https://prodxmon-pa.googleapis.com/v1:insert"
	tsmonFlags.Credentials = credsFile
	tsmonFlags.Target.TargetType = target.TaskType
	tsmonFlags.Target.TaskServiceName = "CTR-DockerOps"
	tsmonFlags.Target.TaskJobName = "CTR-DockerOps"
	tsmonFlags.Flush = "auto"

	// Initialize the library once on application start:
	if err := tsmon.InitializeFromFlags(ctx, &tsmonFlags); err != nil {
		return fmt.Errorf("metrics: error setup tsmon: %s", err)
	}
	return nil
}

// metricsShutdown stops the metrics.
func metricsShutdown(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	log.Printf("Shutting down metrics...")
	tsmon.Shutdown(ctx)
}
