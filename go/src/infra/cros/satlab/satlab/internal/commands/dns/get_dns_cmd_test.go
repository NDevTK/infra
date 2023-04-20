// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dns

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRunGetCmdInjected(t *testing.T) {
	t.Parallel()

	fakeContents := func() (string, error) {
		return "content", nil
	}
	out := new(bytes.Buffer)
	expectedOut := "Satlab internal DNS:\ncontent\n"
	cmd := getDNSRun{}

	cmd.runCmdInjected(out, fakeContents)

	if diff := cmp.Diff(out.String(), expectedOut); diff != "" {
		t.Errorf("unexpected diff in output: %s", diff)
	}
}
