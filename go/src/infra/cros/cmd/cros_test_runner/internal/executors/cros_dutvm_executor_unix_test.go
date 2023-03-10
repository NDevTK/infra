// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build unix
// +build unix

package executors

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

func TestCrosDutVmExecutor_StartCrosDut(t *testing.T) {
	t.Parallel()

	getCmd := func(exec interfaces.ExecutorInterface) *commands.DutServiceStartCmd {
		cmd := commands.NewDutServiceStartCmd(exec)
		cmd.CacheServerAddress = &labapi.IpEndpoint{Address: "localhost", Port: 8080}
		cmd.DutSshAddress = &labapi.IpEndpoint{Address: "1.2.3.4", Port: 22}
		return cmd
	}

	Convey("StartCrosDut success", t, func() {
		expected := &labapi.IpEndpoint{Address: "localhost", Port: 44355}
		ctx := context.Background()
		exec := buildCrosDutVmExecutor()
		cmd := getCmd(exec)
		exec.Container = &mockContainerApi{
			process: func(ctx context.Context, template *api.Template) (string, error) {
				So(template, ShouldNotBeNil)
				So(template.GetCrosDut().GetCacheServer().GetAddress(), ShouldNotBeNil)
				So(template.GetCrosDut().GetCacheServer().GetAddress(), ShouldNotEqual, "localhost")
				So(template.GetCrosDut().GetCacheServer().GetPort(), ShouldEqual, 8080)
				return "localhost:44355", nil
			},
		}

		err := exec.ExecuteCommand(ctx, cmd)

		So(err, ShouldBeNil)
		So(cmd.DutServerAddress, ShouldResemble, expected)
	})

	Convey("StartCrosDut error", t, func() {
		ctx := context.Background()
		exec := buildCrosDutVmExecutor()
		exec.Container = &mockContainerApi{
			process: func(ctx context.Context, template *api.Template) (string, error) {
				return "", fmt.Errorf("ctr error")
			},
		}
		cmd := getCmd(exec)

		err := exec.ExecuteCommand(ctx, cmd)

		So(err, ShouldNotBeNil)
	})
}
