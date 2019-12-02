// Copyright 2019 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"

	. "github.com/smartystreets/goconvey/convey"
)

func newTaskArgs() (ctx context.Context, uid, logdogPrefix string) {
	return context.Background(), "username", "username_at_example.com/12345"
}

func readTestFixture(fixtureBaseName string) *JobDefinition {
	data, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.json", fixtureBaseName))
	So(err, ShouldBeNil)

	req := &swarming.SwarmingRpcsNewTaskRequest{}
	So(json.NewDecoder(bytes.NewReader(data)).Decode(req), ShouldBeNil)

	jd, err := JobDefinitionFromNewTaskRequest(
		req, "test_name", "swarming.example.com")
	So(err, ShouldBeNil)
	So(jd, ShouldNotBeNil)
	return jd
}

func TestGetSwarmRaw(t *testing.T) {
	t.Parallel()

	Convey(`consume non-buildbucket swarming task`, t, func() {
		jd := readTestFixture("raw_swarming_request")

		So(jd.GetSwarming(), ShouldNotBeNil)
		So(jd.SwarmingHostname(), ShouldEqual, "swarming.example.com")
		So(jd.TaskName(), ShouldEqual, "led: test_name")
	})
}

func TestGetKitchenBuild(t *testing.T) {
	t.Parallel()

	Convey(`consume kitchen buildbucket swarming task`, t, func() {
		jd := readTestFixture("raw_kitchen_request")

		So(jd.GetBuildbucket(), ShouldNotBeNil)
		So(jd.SwarmingHostname(), ShouldEqual, "chromium-swarm.appspot.com")
		So(jd.TaskName(), ShouldEqual, "led: test_name")
	})
}

func TestGetBBAgentBuild(t *testing.T) {
	t.Parallel()

	Convey(`consume kitchen buildbucket swarming task`, t, func() {
		jd := readTestFixture("raw_bbagent_request")

		So(jd.GetBuildbucket(), ShouldNotBeNil)
		So(jd.SwarmingHostname(), ShouldEqual, "chromium-swarm-dev.appspot.com")
		So(jd.TaskName(), ShouldEqual, "led: test_name")
	})
}
