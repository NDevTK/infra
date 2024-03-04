// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/openid"
	"go.chromium.org/luci/server/module"
	"go.chromium.org/luci/server/secrets"

	"infra/device_manager/internal/database"
	"infra/device_manager/internal/frontend"
)

func main() {
	modules := []module.Module{
		secrets.NewModuleFromFlags(),
	}

	dbHost := flag.String(
		"db-host",
		"device_manager_db",
		"The DB host location to connect to.",
	)

	dbPort := flag.String(
		"db-port",
		"5432",
		"The DB port number to connect to.",
	)

	dbName := flag.String(
		"db-name",
		"device_manager_db",
		"The DB name to connect to.",
	)

	dbUser := flag.String(
		"db-user",
		"postgres",
		"The DB user to connect as.",
	)

	dbPasswordSecret := flag.String(
		"db-password-secret",
		"devsecret-text://password",
		"The DB password location for Secret Store to use.",
	)

	server.Main(nil, modules, func(srv *server.Server) error {
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

		logging.Infof(srv.Context, "main: installing Services.")

		deviceLeaseServer := frontend.NewServer()
		dbConfig := database.DatabaseConfig{
			DBHost:           *dbHost,
			DBPort:           *dbPort,
			DBName:           *dbName,
			DBUser:           *dbUser,
			DBPasswordSecret: *dbPasswordSecret,
		}
		frontend.SetDBConfig(deviceLeaseServer, dbConfig)
		frontend.InstallServices(deviceLeaseServer, srv)

		logging.Infof(srv.Context, "main: initialization finished")

		return nil
	})
}