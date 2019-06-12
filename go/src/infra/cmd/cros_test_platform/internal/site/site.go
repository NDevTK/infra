// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package site contains functions and constants related to execution of this
// tool in specific environments (e.g., developer workstation vs buildbucket
// build)
package site

import (
	"os"
	"path/filepath"

	"go.chromium.org/luci/auth"
)

// DefaultAuthOptions is an auth.Options struct prefilled with chrome-infra
// defaults.
var DefaultAuthOptions = auth.Options{
	// DO NOT SUBMIT. We're not going to triple copy these ClientID and
	// ClientSecret that originate from the swarming client.
	// In fact,  we probably only need these for developer invocation. The
	// buildbucket build should use a different path entirely.

	// Note that ClientSecret is not really a secret since it's hardcoded into
	// the source code (and binaries). It's totally fine, as long as it's callback
	// URI is configured to be 'localhost'.
	ClientID:     "446450136466-2hr92jrq8e6i4tnsa56b52vacp7t3936.apps.googleusercontent.com",
	ClientSecret: "uBfbay2KCy9t4QveJ-dOqHtp",
	SecretsDir:   secretsDir(),
	Scopes:       []string{auth.OAuthScopeEmail},
}

// SecretsDir returns an absolute path to a directory (in $HOME) to keep secret
// files in (e.g. OAuth refresh tokens) or an empty string if $HOME can't be
// determined.
func secretsDir() string {
	configDir := os.Getenv("XDG_CACHE_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(configDir, "cros_test_platform", "auth")
}
