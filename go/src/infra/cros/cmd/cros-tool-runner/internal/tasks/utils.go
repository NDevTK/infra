// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	build_api "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/lucictx"

	"infra/cros/cmd/cros-tool-runner/internal/common"
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

func findContainer(cm *build_api.ContainerMetadata, lookupKey, name string) *build_api.ContainerImageInfo {
	containers := cm.GetContainers()
	if containers == nil {
		return nil
	}
	imageMap, ok := containers[lookupKey]
	if !ok {
		log.Printf("Image %q not found", name)
		return nil
	}
	return imageMap.Images[name]
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

// dockerAuth will run the gcloud auth cmd and return the token given.
func dockerAuth(ctx context.Context, keyfile string) (string, error) {
	// If keyfile does not exist, we assume that auth is not required.
	// This case is necessary for CTP to run CTF where CTP bot has valid account
	// to pull images.

	var err error
	for i := 0; i < 2; i++ {
		if i > 0 {
			log.Println("retrying after error:", err)
			time.Sleep(5 * time.Second)
		}
		err = activateAccount(ctx, keyfile)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", fmt.Errorf("could not activate account last error: %s", err)
	}

	cmd := exec.Command("gcloud", "auth", "print-access-token")
	out, _, err := common.RunWithTimeout(ctx, cmd, time.Minute, true)
	if err != nil {
		return "", errors.Annotate(err, "failed running 'gcloud auth print-access-token'").Err()
	}
	return out, nil
}

func activateAccount(ctx context.Context, keyfile string) error {
	if _, err := os.Stat(keyfile); err == nil {
		// keyfile exists
		cmd := exec.Command("gcloud", "auth", "activate-service-account",
			fmt.Sprintf("--key-file=%v", keyfile))
		out, stderr, err := common.RunWithTimeout(ctx, cmd, time.Minute, true)
		if err != nil {
			log.Printf("Failed running gcloud auth: %s\n%s", err, stderr)
			return errors.Annotate(err, "gcloud auth").Err()
		}
		log.Printf("gcloud auth done. Result: %s", out)
	} else if os.IsNotExist(err) {
		// keyfile doesn't exist.
		// For this case, we will assume that env has account with proper permissions.
		log.Printf("Skipping gcloud auth as keyfile does not exist")
	} else {
		// keyfile may or may not exist. See err for details.
		return errors.Annotate(err, "error with keyfile").Err()
	}
	return nil

}
