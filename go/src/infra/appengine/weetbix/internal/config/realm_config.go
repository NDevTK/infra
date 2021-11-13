// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"context"
	"fmt"

	"go.chromium.org/luci/server/auth/realms"
)

func Realm(ctx context.Context, realm string) (*RealmConfig, error) {
	project, shortRealm := realms.Split(realm)
	pc, err := Project(ctx, project)
	if err != nil {
		return nil, err
	}
	for _, rc := range pc.GetRealms() {
		if rc.Name == shortRealm {
			return rc, nil
		}
	}
	return nil, fmt.Errorf("not found config for realm %s", realm)
}
