// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/openid"
	"google.golang.org/grpc"

	"infra/device_manager/internal/frontend"
)

// InstallServices takes a DeviceLeaseServiceServer and exposes it to a LUCI
// prpc.Server.
func InstallServices(s *frontend.Server, srv grpc.ServiceRegistrar) {
	api.RegisterDeviceLeaseServiceServer(srv, s)
}

func main() {
	server.Main(nil, nil, func(srv *server.Server) error {
		logging.Infof(srv.Context, "main: initializing server")

		// This allows auth to use Identity tokens.
		srv.SetRPCAuthMethods([]auth.Method{
			// The primary authentication method.
			&openid.GoogleIDTokenAuthMethod{
				AudienceCheck: openid.AudienceMatchesHost,
				SkipNonJWT:    true, // pass OAuth2 access tokens through
			},
			// Backward compatibility for RPC Explorer and old clients.
			&auth.GoogleOAuth2Method{
				Scopes: []string{"https://www.googleapis.com/auth/userinfo.email"},
			},
		})

		logging.Infof(srv.Context, "Installing Services.")
		InstallServices(frontend.NewServer(), srv)

		logging.Infof(srv.Context, "main: initialization finished")
		return nil
	})
}
