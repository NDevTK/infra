// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executors

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"infra/cros/cmd/cros_test_runner/internal/commands"
	"infra/cros/cmd/cros_test_runner/internal/containers"
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
)

func TestCacheServerExecutor_StartCacheServer(t *testing.T) {
	t.Parallel()

	getCmd := func(exec interfaces.ExecutorInterface) *commands.DutVmCacheServerStartCmd {
		cmd := commands.NewDutVmCacheServerStartCmd(exec)
		duts := []*labapi.Dut{{
			Id: &labapi.Dut_Id{Value: "VM"},
			DutType: &labapi.Dut_Chromeos{
				Chromeos: &labapi.Dut_ChromeOS{},
			}}}
		cmd.DutTopology = &labapi.DutTopology{
			Duts: duts,
		}
		return cmd
	}

	Convey("StartCacheServer success", t, func() {
		expected := &labapi.IpEndpoint{Address: "4.3.2.1", Port: 8080}
		ctx := context.Background()
		exec := buildCacheServerExecutor()
		cmd := getCmd(exec)
		exec.Container = &mockContainerApi{
			process: func(ctx context.Context, template *api.Template) (string, error) {
				So(template, ShouldNotBeNil)
				return "4.3.2.1:8080", nil
			},
		}

		err := exec.ExecuteCommand(ctx, cmd)

		So(err, ShouldBeNil)
		So(cmd.CacheServerAddress, ShouldResemble, expected)
	})

	Convey("StartCacheServer error", t, func() {
		ctx := context.Background()
		exec := buildCacheServerExecutor()
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

func buildCacheServerExecutor() *CacheServerExecutor {
	ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
	ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
	cont := containers.NewCacheServerTemplatedContainer("container/image/path", ctr)
	exec := NewCacheServerExecutor(cont)
	return exec
}
