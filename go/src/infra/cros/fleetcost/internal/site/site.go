// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package site

import (
	"os"
	"path/filepath"
	"strings"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/hardcoded/chromeinfra"
)

// progName is the name of the program
const progName = "fleetcost"

// DefaultAuthOptions is an auth.Options struct prefilled with chrome-infra
// defaults.
var DefaultAuthOptions = chromeinfra.SetDefaultAuthOptions(auth.Options{
	Scopes:     GetAuthScopes(DefaultAuthScopes),
	SecretsDir: SecretsDir(),
})

// GetAuthScopes get environment scopes if set
// Otherwise, return default scopes
func GetAuthScopes(defaultScopes []string) []string {
	e := os.Getenv("OAUTH_SCOPES")
	if e != "" {
		return strings.Split(e, "|")
	}
	return defaultScopes
}

// SecretsDir customizes the location for auth-related secrets.
func SecretsDir() string {
	configDir := os.Getenv("XDG_CACHE_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(configDir, progName, "auth")
}

// DefaultAuthScopes is the default scopes for shivas login
var DefaultAuthScopes = []string{auth.OAuthScopeEmail, "https://www.googleapis.com/auth/spreadsheets"}
