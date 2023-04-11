// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"infra/cmd/shivas/site"
	"infra/libs/skylab/buildbucket"
)

// TestScheduleRepairBuilder tests that scheduling a repair builder produces the correct
// taskID and the right URL. This test does NOT emulate the buildbucket client on a deep level.
func TestScheduleRepairBuilder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &fakeClient{}
	taskURL, err := scheduleRepairBuilder(ctx, client, site.Environment{}, "fake-labstation1", true, true, true, "admin-session:bla bla")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	expected := "https://ci.chromium.org/p/chromeos/builders/labpack/labpack/b1"
	actual := taskURL
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected diff: %s", diff)
	}
}

// FakeClient is a fake buildbucket client.
type fakeClient struct{}

// ScheduleLabpackTask is a fake method that returns a fixed buildbucket ID of 1.
func (c *fakeClient) ScheduleLabpackTask(ctx context.Context, _ *buildbucket.ScheduleLabpackTaskParams, _ string) (string, int64, error) {
	return fmt.Sprintf(buildbucket.BuildURLFmt, "chromeos", "labpack", "labpack", 1), 1, nil
}

// BuildURL returns the build URL.
func (c *fakeClient) BuildURL(buildID int64) string {
	return fmt.Sprintf(buildbucket.BuildURLFmt, "chromeos", "labpack", "labpack", buildID)
}
