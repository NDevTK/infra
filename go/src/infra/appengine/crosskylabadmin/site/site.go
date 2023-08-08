// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package site

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/gcloud/gs"
	"go.chromium.org/luci/common/retry"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/hardcoded/chromeinfra"
)

// ProdUFS is the unified fleet service.
const ProdUFS = "ufs.api.cr.dev"

// StagingService is the staging instance of CrOSSkylabAdmin.
const StagingService = "skylab-staging-bot-fleet.appspot.com"

// ProdService is the prod instance of CrOSSkylabAdmin.
const ProdService = "chromeos-skylab-bot-fleet.appspot.com"

// LocalService is the local instance of CrOSkylabAdmin.
const LocalService = "127.0.0.1:8000"

// DefaultPRPCOptions is used for PRPC clients.  If it is nil, the
// default value is used.  See prpc.Options for details.
//
// This is provided so it can be overridden for testing.
var DefaultPRPCOptions = prpcOptionWithUserAgent(fmt.Sprintf("adminclient/%s", VersionNumber))

// DefaultAuthOptions is an auth.Options struct prefilled with chrome-infra
// defaults.
var DefaultAuthOptions = chromeinfra.SetDefaultAuthOptions(auth.Options{
	Scopes:     append(gs.ReadWriteScopes, auth.OAuthScopeEmail),
	SecretsDir: SecretsDir(),
})

// SecretsDir customizes the location for auth-related secrets.
func SecretsDir() string {
	configDir := os.Getenv("XDG_CACHE_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(configDir, "adminclient", "auth")
}

// VersionNumber is the version number for the tool. It follows the Semantic
// Versioning Specification (http://semver.org) and the format is:
// "MAJOR.MINOR.0+BUILD_TIME".
// We can ignore the PATCH part (i.e. it's always 0) to make the maintenance
// work easier.
// We can also print out the build time (e.g. 20060102150405) as the METADATA
// when show version to users.
var VersionNumber = fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)

// Major is the Major version number
// See b:215410393 for context behind bumping the version.
const Major = 4

// Minor is the Minor version number
const Minor = 0

// Patch is the Patch version number
const Patch = 0

// prpcOptionWithUserAgent create prpc option with custom UserAgent.
//
// DefaultOptions provides Retry ability in case we have issue with service.
//
// There are many permanent failures that are marked as tranisent. In order to
// mitigate the user-facing impacts of this when developing or testing things,
// we perform the retries immediately instead of using exponential backoff.
func prpcOptionWithUserAgent(userAgent string) *prpc.Options {
	return &prpc.Options{
		UserAgent: userAgent,
		Retry: func() retry.Iterator {
			return &retry.Limited{
				Delay:   10 * time.Millisecond,
				Retries: 2,
			}
		},
	}
}
