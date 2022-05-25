// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.chromium.org/luci/gae/service/datastore"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/frontend/datastore/stableversion/satlab"
)

// TestSetSatlabStableVersion tests that SetSatlabStableVersion returns a not-yet-implemented response.
func TestSetSatlabStableVersion(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	ctx = withSplitInventory(ctx)
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()

	_, sErr := tf.Inventory.SetSatlabStableVersion(ctx, &fleet.SetSatlabStableVersionRequest{
		Strategy: &fleet.SetSatlabStableVersionRequest_SatlabHostnameStrategy{
			SatlabHostnameStrategy: &fleet.SatlabHostnameStrategy{
				Hostname: "satlab-host1",
			},
		},
		CrosVersion:     "R12-1234.56.78",
		FirmwareVersion: "Google_Something.1234.56.0",
		FirmwareImage:   "something-firmware/R12-1234.56.78",
	})
	if sErr != nil {
		t.Errorf("unexpected error when inserting record: %s", sErr)
	}

	expected := &satlab.SatlabStableVersionEntry{
		ID:        "satlab-host1",
		OS:        "R12-1234.56.78",
		FW:        "Google_Something.1234.56.0",
		FWImage:   "something-firmware/R12-1234.56.78",
		Base64Req: "Gg5SMTItMTIzNC41Ni43OCIaR29vZ2xlX1NvbWV0aGluZy4xMjM0LjU2LjAqIXNvbWV0aGluZy1maXJtd2FyZS9SMTItMTIzNC41Ni43OBIOEgxzYXRsYWItaG9zdDE=",
	}
	actual, _ := satlab.GetSatlabStableVersionEntryByRawID(ctx, "satlab-host1")
	if diff := cmp.Diff(expected, actual, cmpopts.IgnoreUnexported(satlab.SatlabStableVersionEntry{})); diff != "" {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}

// TestDeleteSatlabStableVersion tests that SetSatlabStableVersion returns a not-yet-implemented response.
func TestDeleteSatlabStableVersion(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	ctx = withSplitInventory(ctx)
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()

	_, sErr := tf.Inventory.SetSatlabStableVersion(ctx, &fleet.SetSatlabStableVersionRequest{
		Strategy: &fleet.SetSatlabStableVersionRequest_SatlabHostnameStrategy{
			SatlabHostnameStrategy: &fleet.SatlabHostnameStrategy{
				Hostname: "satlab-host1",
			},
		},
		CrosVersion:     "R12-1234.56.78",
		FirmwareVersion: "Google_Something.1234.56.0",
		FirmwareImage:   "something-firmware/R12-1234.56.78",
	})
	if sErr != nil {
		t.Errorf("unexpected error when inserting record: %s", sErr)
	}

	_, gErr := satlab.GetSatlabStableVersionEntryByRawID(ctx, "satlab-host1")
	if gErr != nil {
		t.Errorf("unexpected error retrieving after insertion: %s", gErr)
	}

	_, dErr := tf.Inventory.DeleteSatlabStableVersion(ctx, &fleet.DeleteSatlabStableVersionRequest{
		Strategy: &fleet.DeleteSatlabStableVersionRequest_SatlabHostnameDeletionCriterion{
			SatlabHostnameDeletionCriterion: &fleet.SatlabHostnameDeletionCriterion{
				Hostname: "satlab-host1",
			},
		},
	})
	if dErr != nil {
		t.Errorf("unexpected error deleting record that is present: %s", dErr)
	}

	_, err := satlab.GetSatlabStableVersionEntryByRawID(ctx, "satlab-host1")
	if err == nil {
		t.Errorf("getting record after deletion should have failed")
	}
	if !datastore.IsErrNoSuchEntity(err) {
		t.Errorf("unexpected error: %s", err)
	}
}
