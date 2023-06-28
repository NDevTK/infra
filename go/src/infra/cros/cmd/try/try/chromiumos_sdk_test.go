// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package try

import (
	"fmt"
	"os"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/cmd"
	bb "infra/cros/lib/buildbucket"
)

// TestChromiumOSSDKGetBuilderFullName tests chromiumosSDKRun.getBuilderFullName.
func TestChromiumOSSDKGetBuilderFullName(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		production          bool
		expectedBuilderName string
	}{
		{true, "chromeos/infra/build-chromiumos-sdk"},
		{false, "chromeos/staging/staging-build-chromiumos-sdk"},
	} {
		run := chromiumOSSDKRun{
			tryRunBase: tryRunBase{
				production: tc.production,
			},
		}
		actualBuilderName := run.getBuilderFullName()
		assert.StringsEqual(t, actualBuilderName, tc.expectedBuilderName)
	}
}

// chromiumOSSDKRunTestConfig contains info for an end-to-end test of chromiumOSSDKRun.Run().
type chromiumOSSDKRunTestConfig struct {
	production      bool
	expectedBucket  string
	expectedBuilder string
}

func doChromiumOSSDKRun(t *testing.T, tc chromiumOSSDKRunTestConfig) {
	t.Helper()

	// Set up properties tempfile.
	propsFile, err := os.CreateTemp("", "input_props")
	defer os.Remove(propsFile.Name())
	assert.NilError(t, err)

	// Set up fake commands.
	cmdRunner := &cmd.FakeCommandRunnerMulti{
		CommandRunners: []cmd.FakeCommandRunner{
			bb.FakeAuthInfoRunner("bb", 0),
			bb.FakeAuthInfoRunner("led", 0),
			bb.FakeAuthInfoRunnerSuccessStdout("led", "sundar@google.com"),
			*fakeLEDGetBuilderRunner(tc.expectedBucket, tc.expectedBuilder, true),
			bb.FakeBBAddRunner(
				[]string{
					"bb",
					"add",
					fmt.Sprintf("%s/%s", tc.expectedBucket, tc.expectedBuilder),
					"-t",
					"tryjob-launcher:sundar@google.com",
					"-p",
					"@" + propsFile.Name(),
				},
				"12345",
			),
		},
	}

	// Set up fake chromiumOSSDKRun.
	run := chromiumOSSDKRun{
		tryRunBase: tryRunBase{
			cmdRunner:            cmdRunner,
			production:           tc.production,
			skipProductionPrompt: true,
		},
		propsFile: propsFile,
	}

	// Try running!
	ret := run.Run(nil, nil, nil)
	assert.IntsEqual(t, ret, Success)
}

// TestChromiumOSSDKRun_Production is an end-to-end test of chromiumOSSDKRun.Run() for a production build.
func TestChromiumOSSDKRun_Production(t *testing.T) {
	t.Parallel()
	tc := chromiumOSSDKRunTestConfig{
		production:      true,
		expectedBucket:  "chromeos/infra",
		expectedBuilder: "build-chromiumos-sdk",
	}
	doChromiumOSSDKRun(t, tc)
}

// TestChromiumOSSDKRun_Staging is an end-to-end test of chromiumOSSDKRun.Run() for a staging build.
func TestChromiumOSSDKRun_Staging(t *testing.T) {
	t.Parallel()
	tc := chromiumOSSDKRunTestConfig{
		production:      false,
		expectedBucket:  "chromeos/staging",
		expectedBuilder: "staging-build-chromiumos-sdk",
	}
	doChromiumOSSDKRun(t, tc)
}
