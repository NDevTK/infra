// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

import (
	"context"
	"fmt"

	"go.chromium.org/luci/auth/identity"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"
	"go.chromium.org/luci/server/auth/realms"
)

func mockUser(ctx context.Context, user string) context.Context {
	return auth.WithState(ctx, &authtest.FakeState{
		Identity: identity.Identity(fmt.Sprintf("user:%s", user)),
	})
}

func mockRealmPerms(ctx context.Context, realm string, permission realms.Permission) {
	state := auth.GetState(ctx).(*authtest.FakeState)
	state.IdentityPermissions = append(state.IdentityPermissions, authtest.RealmPermission{
		Realm:      realm,
		Permission: permission,
	})
}
