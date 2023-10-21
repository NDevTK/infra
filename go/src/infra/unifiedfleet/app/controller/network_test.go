// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"fmt"
	"testing"

	ufspb "infra/unifiedfleet/api/v1/models"

	"github.com/google/go-cmp/cmp"
)

// TestGetFreeIPSimple tests getting a free IP address form a mostly empty Vlan.
func TestGetFreeIPSimple(t *testing.T) {
	t.Parallel()

	ctx := testingContext()

	const originalIP = "192.64.0.0"
	// We reserve 192.64.0.0 unconditionally,
	// followed by the first 10 addresses 192.64.0.{1..10}
	// Thus the first address that's free for our use is 192.64.0.11
	const firstFreeIP = "192.64.0.11"

	_, err := CreateVlan(ctx, &ufspb.Vlan{
		Name:        "fake-vlan",
		VlanAddress: fmt.Sprintf("%s/24", originalIP),
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := getFreeIP(ctx, "fake-vlan", 1)
	if err != nil {
		t.Error(err)
	}
	if len(res) != 1 {
		t.Errorf("res has bad length %d: %#v", len(res), res)
	}

	if diff := cmp.Diff(res[0].GetIpv4Str(), firstFreeIP); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}
