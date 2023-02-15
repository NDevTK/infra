// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package site contains site local constants for the crosfleet tool.
package site

import (
	"fmt"
	"os"
	"path/filepath"

	"go.chromium.org/luci/auth"
	buildbucket_pb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/hardcoded/chromeinfra"
)

// Environment contains environment specific values.
type Environment struct {
	SwarmingService string
	UFSService      string

	// Buildbucket-specific values.
	BuildbucketService string
	DefaultCTPBuilder  *buildbucket_pb.BuilderID
	DUTLeaserBuilder   *buildbucket_pb.BuilderID
}

// Prod is the environment for prod.
var Prod = Environment{
	SwarmingService: "https://chromeos-swarming.appspot.com/",
	UFSService:      "ufs.api.cr.dev",

	BuildbucketService: "cr-buildbucket.appspot.com",
	DefaultCTPBuilder: &buildbucket_pb.BuilderID{
		Project: "chromeos",
		Bucket:  "testplatform",
		Builder: "cros_test_platform",
	},
	DUTLeaserBuilder: &buildbucket_pb.BuilderID{
		Project: "chromeos",
		Bucket:  "test_runner",
		Builder: "dut_leaser",
	},
}

// Dev is the environment for dev.
var Dev = Environment{
	SwarmingService: "https://chromeos-swarming.appspot.com/",
	UFSService:      "staging.ufs.api.cr.dev",

	BuildbucketService: "cr-buildbucket.appspot.com",
	DefaultCTPBuilder: &buildbucket_pb.BuilderID{
		Project: "chromeos",
		Bucket:  "testplatform",
		Builder: "cros_test_platform-dev",
	},
	DUTLeaserBuilder: &buildbucket_pb.BuilderID{
		Project: "chromeos",
		Bucket:  "test_runner",
		Builder: "dut_leaser",
	},
}

// DefaultAuthOptions is an auth.Options struct prefilled with chrome-infra
// defaults.
var DefaultAuthOptions = auth.Options{
	// Note that ClientSecret is not really a secret since it's hardcoded into
	// the source code (and binaries). It's totally fine, as long as it's callback
	// URI is configured to be 'localhost'. If someone decides to reuse such
	// ClientSecret they have to run something on user's local machine anyway
	// to get the refresh_token.
	ClientID:          "446450136466-mj75ourhccki9fffaq8bc1e50di315po.apps.googleusercontent.com",
	ClientSecret:      "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",
	LoginSessionsHost: chromeinfra.LoginSessionsHost,
	SecretsDir:        SecretsDir(),
	Scopes:            []string{auth.OAuthScopeEmail, gitiles.OAuthScope},
}

// VersionNumber is the service version number for the crosfleet tool.
const VersionNumber = 4

// DefaultPRPCOptions is used for PRPC clients.  If it is nil, the
// default value is used.  See prpc.Options for details.
//
// This is provided so it can be overridden for testing.
var DefaultPRPCOptions = prpcOptionWithUserAgent(fmt.Sprintf("crosfleet/%d", VersionNumber))

// SecretsDir returns an absolute path to a directory (in $HOME) to keep secret
// files in (e.g. OAuth refresh tokens) or an empty string if $HOME can't be
// determined (happens in some degenerate cases, it just disables auth token
// cache).
func SecretsDir() string {
	configDir := os.Getenv("XDG_CACHE_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(configDir, "crosfleet", "auth")
}

// prpcOptionWithUserAgent create prpc option with custom UserAgent.
//
// DefaultOptions provides Retry ability in case we have issue with service.
func prpcOptionWithUserAgent(userAgent string) *prpc.Options {
	options := *prpc.DefaultOptions()
	options.UserAgent = userAgent
	return &options
}
