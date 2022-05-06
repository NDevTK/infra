// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package utils contains various helper functions.
package utils

import (
	"context"
	"strings"

	"go.chromium.org/luci/grpc/appstatus"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/realms"
	"google.golang.org/grpc/codes"
)

// SplitRealm splits the Realm into the LUCI project name and the (sub)Realm.
// Returns empty strings if the provided Realm doesn't have a valid format.
func SplitRealm(realm string) (proj string, subRealm string) {
	parts := strings.SplitN(realm, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// HasPermissions is a wrapper around luci/server/auth.HasPermission that checks
// whether the user has all the listed permissions and return an appstatus
// annotated error if users have no permission.
func HasPermissions(ctx context.Context, permissions []realms.Permission, realm string, attrs realms.Attrs) error {
	for _, perm := range permissions {
		allowed, err := auth.HasPermission(ctx, perm, realm, nil)
		if err != nil {
			return err
		}
		if !allowed {
			return appstatus.Errorf(codes.PermissionDenied, `caller does not have permission %s in realm %q`, perm, realm)
		}
	}
	return nil
}
