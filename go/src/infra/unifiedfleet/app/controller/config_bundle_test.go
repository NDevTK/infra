// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"testing"

	"go.chromium.org/chromiumos/config/go/api"
	"go.chromium.org/chromiumos/config/go/payload"
	"google.golang.org/protobuf/proto"
)

func mockConfigBundle(id string, programId string, name string) *payload.ConfigBundle {
	return &payload.ConfigBundle{
		DesignList: []*api.Design{
			{
				Id: &api.DesignId{
					Value: id,
				},
				ProgramId: &api.ProgramId{
					Value: programId,
				},
				Name: name,
			},
		},
	}
}

func TestUpdateConfigBundle(t *testing.T) {
	t.Parallel()
	ctx := testingContext()

	t.Run("update non-existent ConfigBundle", func(t *testing.T) {
		cb1 := mockConfigBundle("design1", "program1", "name1")
		cbBytes, err := proto.Marshal(cb1)
		if err != nil {
			t.Fatalf("UpdateConfigBundle failed: %s", err)
		}

		gotBytes, err := UpdateConfigBundle(ctx, cbBytes, nil, true)
		if err != nil {
			t.Fatalf("UpdateConfigBundle failed: %s", err)
		}

		got := &payload.ConfigBundle{}
		if err := proto.Unmarshal(gotBytes, got); err != nil {
			t.Fatalf("UpdateConfigBundle failed to unmarshal ConfigBundle bytes: %s", err)
		}
		if !proto.Equal(cb1, got) {
			t.Errorf("UpdateConfigBundle returned unexpected diff (-want +got):\n%s\n%s", cb1, got)
		}
	})

	t.Run("update existent ConfigBundle", func(t *testing.T) {
		cb2 := mockConfigBundle("design2", "program2", "name2")
		cb2update := mockConfigBundle("design2", "program2", "name2update")

		// Insert cb2 into datastore
		cb2Bytes, err := proto.Marshal(cb2)
		if err != nil {
			t.Fatalf("UpdateConfigBundle failed: %s", err)
		}
		_, _ = UpdateConfigBundle(ctx, cb2Bytes, nil, true)

		// Update cb2
		cb2updateBytes, err := proto.Marshal(cb2update)
		if err != nil {
			t.Fatalf("UpdateConfigBundle failed: %s", err)
		}
		gotBytes, err := UpdateConfigBundle(ctx, cb2updateBytes, nil, true)
		if err != nil {
			t.Fatalf("UpdateConfigBundle failed: %s", err)
		}

		got := &payload.ConfigBundle{}
		if err := proto.Unmarshal(gotBytes, got); err != nil {
			t.Fatalf("UpdateConfigBundle failed to unmarshal ConfigBundle bytes: %s", err)
		}
		if !proto.Equal(cb2update, got) {
			t.Errorf("UpdateConfigBundle returned unexpected diff (-want +got):\n%s\n%s", cb2update, got)
		}
	})

	t.Run("update ConfigBundle with invalid IDs", func(t *testing.T) {
		cb3 := mockConfigBundle("", "", "")
		cb3Bytes, err := proto.Marshal(cb3)
		if err != nil {
			t.Fatalf("UpdateConfigBundle failed: %s", err)
		}

		gotBytes, err := UpdateConfigBundle(ctx, cb3Bytes, nil, true)
		if err == nil {
			t.Errorf("UpdateConfigBundle succeeded with empty IDs")
		}

		got := &payload.ConfigBundle{}
		if err := proto.Unmarshal(gotBytes, got); err != nil {
			t.Fatalf("UpdateConfigBundle failed to unmarshal ConfigBundle bytes: %s", err)
		}

		cbNil := &payload.ConfigBundle{}
		if !proto.Equal(cbNil, got) {
			t.Errorf("UpdateConfigBundle returned unexpected diff (-want +got):\n%s\n%s", cbNil, got)
		}
	})
}
