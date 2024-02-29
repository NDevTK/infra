// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"context"
	"fmt"
	"os"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/secrets"
)

// GetSecret gets the active secret using the LUCI Secrets package.
//
// The string must be a base64 encoded string. It should not have padding and be
// in raw encoded form.
func GetSecret(ctx context.Context, secretLoc string) (string, error) {
	secret, err := secrets.StoredSecret(ctx, secretLoc)
	if err != nil {
		logging.Errorf(ctx, "GetSecret: failed to get secret %s: %s", secretLoc, err)
		return "", err
	}
	return string(secret.Active), nil
}

// GetEnvVar tries to get the corresponding environment variable for a string.
func GetEnvVar(ctx context.Context, k string) (string, error) {
	v := os.Getenv(k)
	if v == "" {
		return "", fmt.Errorf("GetEnvVar: %s environment variable not set", k)
	}
	return v, nil
}
