// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"testing"

	"infra/appengine/crosskylabadmin/internal/app/frontend/routing"
)

func TestRouteAuditTaskImpl(t *testing.T) {
	if taskType, _ := routeAuditTaskImpl(nil); taskType != routing.Legacy {
		t.Errorf("route audit task should always return legacy")
	}
}
