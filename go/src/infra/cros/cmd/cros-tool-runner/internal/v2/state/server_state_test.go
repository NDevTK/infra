// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package state

import (
	"testing"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func TestServerState(t *testing.T) {
	ServerState.TemplateRequest.Add("1",
		&api.StartTemplatedContainerRequest{
			Name: "name1",
		})
	exist := ServerState.TemplateRequest.Exist("1")
	if !exist {
		t.Fatalf("exist should be true")
	}
}
