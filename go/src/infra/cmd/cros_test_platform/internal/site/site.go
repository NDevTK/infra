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
	"go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/gcloud/gs"
	"go.chromium.org/luci/hardcoded/chromeinfra"
)

// DefaultAuthOptions is an auth.Options struct prefilled with command-wide
// defaults.
//
// These defaults support invocation of the command in developer environments.
// The recipe invodation in a BuildBucket should override these defaults.
var DefaultAuthOptions = auth.Options{
	// Note that ClientSecret is not really a secret since it's hardcoded into
	// the source code (and binaries). It's totally fine, as long as it's callback
	// URI is configured to be 'localhost'.
	ClientID:          "446450136466-mj75ourhccki9fffaq8bc1e50di315po.apps.googleusercontent.com",
	ClientSecret:      "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",
	LoginSessionsHost: chromeinfra.LoginSessionsHost,
	SecretsDir:        secretsDir(),
	Scopes:            append(gs.ReadOnlyScopes, gitiles.OAuthScope, auth.OAuthScopeEmail),
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
