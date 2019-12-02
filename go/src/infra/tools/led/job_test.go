// Copyright 2019 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSwarmingToSwarming(t *testing.T) {
	t.Parallel()

	Convey(`consume non-buildbucket swarming task`, t, func() {
		jd := readTestFixture("raw_swarming_request")

		ntr, err := jd.ToSwarmingNewTask("username")
		So(err, ShouldBeNil)
		So(ntr, ShouldNotBeNil)
	})
}
