// Copyright 2023 The ChromiumOS Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package cli

import (
	"net/http"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"golang.org/x/oauth2"
)

var (
	authOptions = chromeinfra.DefaultAuthOptions()
	authFlags   authcli.Flags
)

// Set the auth scopes. Must be called before Init.
func SetAuthScopes(scopes ...string) {
	assertInited(false)
	authOptions.Scopes = scopes
}

// Return an authenticated HTTP client. Must be called after Init.
func AuthenticatedHTTPClient() (*http.Client, error) {
	assertInited(true)
	authOptions, err := authFlags.Options()
	if err != nil {
		return nil, err
	}
	authenticator := auth.NewAuthenticator(Context, auth.SilentLogin, authOptions)
	return authenticator.Client()
}

// Return an authenticated TokenSource. Must be called after Init.
func AuthenticatedTokenSource() (oauth2.TokenSource, error) {
	assertInited(true)
	authOptions, err := authFlags.Options()
	if err != nil {
		return nil, err
	}
	authenticator := auth.NewAuthenticator(Context, auth.SilentLogin, authOptions)
	return authenticator.TokenSource()
}
