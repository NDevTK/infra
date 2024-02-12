// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package try

import (
	"context"
	"fmt"
	"os"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/cmd"
	bb "infra/cros/lib/buildbucket"

	bapipb "go.chromium.org/chromiumos/infra/proto/go/chromite/api"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestValidate_createPreMPKeysRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Test the good workflow
	cmdRunner := fakeBBBuildersRunner("chromeos/staging", []string{"staging-key-manager"})
	f := createPreMPKeysRun{
		tryRunBase: tryRunBase{
			cmdRunner: cmdRunner,
			bbClient:  bb.NewClient(cmdRunner, nil, nil),
		},
		buildTarget: "atlas",
	}
	assert.NilError(t, f.validate(ctx))

	// No build target provided.
	f.buildTarget = ""
	assert.NonNilError(t, f.validate(ctx))
}

type createPreMPKeysTestConfig struct {
	buildTarget string
	dryrun      bool
}

func doCreatePreMPKeysTest(t *testing.T, tc *createPreMPKeysTestConfig) {
	t.Helper()
	propsFile, err := os.CreateTemp("", "input_props")
	defer os.Remove(propsFile.Name())
	assert.NilError(t, err)

	f := &cmd.FakeCommandRunnerMulti{
		CommandRunners: []cmd.FakeCommandRunner{
			bb.FakeAuthInfoRunner("bb", 0),
			bb.FakeAuthInfoRunner("led", 0),
			bb.FakeAuthInfoRunnerSuccessStdout("led", "sundar@google.com"),
		},
	}
	expectedBucket := "chromeos/staging"
	expectedBuilder := "staging-key-manager"
	f.CommandRunners = append(
		f.CommandRunners,
		*fakeLEDGetBuilderRunner(expectedBucket, expectedBuilder, true),
	)
	expectedAddCmd := []string{"bb", "add", fmt.Sprintf("%s/%s", expectedBucket, expectedBuilder)}
	expectedAddCmd = append(expectedAddCmd, "-t", "tryjob-launcher:sundar@google.com")

	expectedAddCmd = append(expectedAddCmd, "-p", fmt.Sprintf("@%s", propsFile.Name()))
	if !tc.dryrun {
		f.CommandRunners = append(f.CommandRunners, bb.FakeBBAddRunner(expectedAddCmd, "12345"))
	}

	r := createPreMPKeysRun{
		tryRunBase: tryRunBase{
			cmdRunner: f,
			dryrun:    tc.dryrun,
		},
		propsFile:   propsFile,
		buildTarget: tc.buildTarget,
	}
	ret := r.Run(nil, nil, nil)
	assert.IntsEqual(t, ret, Success)

	properties, err := bb.ReadStructFromFile(propsFile.Name())
	assert.NilError(t, err)

	// Check that the requests are populated correctly.
	createPreMPKeysRequests := properties.GetFields()["create_premp_keys_requests"].GetListValue()
	jsonRequest, err := createPreMPKeysRequests.GetValues()[0].GetStructValue().MarshalJSON()
	assert.NilError(t, err)
	var createPreMPKeysRequest bapipb.CreatePreMPKeysRequest
	err = protojson.Unmarshal([]byte(jsonRequest), &createPreMPKeysRequest)
	assert.NilError(t, err)

	assert.StringsEqual(
		t,
		createPreMPKeysRequest.BuildTarget.Name,
		tc.buildTarget,
	)
}

func TestCreatePreMPKeys_dryrun(t *testing.T) {
	t.Parallel()
	doCreatePreMPKeysTest(t, &createPreMPKeysTestConfig{
		buildTarget: "atlas",
		dryrun:      true,
	})
}

func TestCreatePreMPKeys_success(t *testing.T) {
	t.Parallel()
	doCreatePreMPKeysTest(t, &createPreMPKeysTestConfig{
		buildTarget: "atlas",
	})
}
