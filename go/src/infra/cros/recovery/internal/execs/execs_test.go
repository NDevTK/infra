// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package execs

import (
	"context"
	"testing"
)

func TestRunExec(t *testing.T) {
	ctx := context.Background()
	const actionExecWrong = "wrong_name"
	const actionExecGood = "sample_pass"
	const actionExecBad = "sample_fail"
	const actionExecMetricsAction = "sample_metrics_action"
	args := &RunArgs{}
	actionArgs := []string{"action", "args"}
	t.Run("Incorrect name", func(t *testing.T) {
		t.Parallel()
		err := Run(ctx, NewExecInfo(args, actionExecWrong, actionArgs, 0, nil))
		if err == nil {
			t.Errorf("Expected to fail")
		}
		if err.Error() != "exec \"wrong_name\": not found" {
			t.Errorf("Did not have expected explanation. Got: %q", err.Error())
		}
	})
	t.Run("Good sample", func(t *testing.T) {
		t.Parallel()
		if err := Run(ctx, NewExecInfo(args, actionExecGood, actionArgs, 0, nil)); err != nil {
			t.Errorf("Expected to pass")
		}
	})
	t.Run("Bad sample", func(t *testing.T) {
		t.Parallel()
		if err := Run(ctx, NewExecInfo(args, actionExecBad, actionArgs, 0, nil)); err == nil {
			t.Errorf("Expected to have status Fail")
		}
	})
	t.Run("Send metrics action", func(t *testing.T) {
		t.Parallel()
		if err := Run(ctx, NewExecInfo(args, actionExecMetricsAction, actionArgs, 0, nil)); err != nil {
			t.Errorf("Expected to pass")
		}
	})
}
