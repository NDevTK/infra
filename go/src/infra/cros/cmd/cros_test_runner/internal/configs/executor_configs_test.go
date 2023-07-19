// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"testing"

	"infra/cros/cmd/common_lib/common_configs"
	"infra/cros/cmd/common_lib/common_executors"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
	"infra/cros/cmd/cros_test_runner/internal/executors"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetExecutor_UnsupportedExecutorType(t *testing.T) {
	t.Parallel()
	Convey("Unsupported executor type", t, func() {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := common_configs.NewContainerConfig(ctr, nil, false)
		execConfig := NewExecutorConfig(ctr, contConfig)
		executor, err := execConfig.GetExecutor(executors.NoExecutorType)
		So(executor, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestGetExecutor_SupportedExecutorType(t *testing.T) {
	t.Parallel()
	Convey("Supported executor type", t, func() {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := common_configs.NewContainerConfig(ctr, getMockContainerImagesInfo(), false)
		execConfig := NewExecutorConfig(ctr, contConfig)

		executor, err := execConfig.GetExecutor(executors.NoExecutorType)
		So(executor, ShouldBeNil)
		So(err, ShouldNotBeNil)

		executor, err = execConfig.GetExecutor(executors.InvServiceExecutorType)
		So(executor, ShouldNotBeNil)
		So(err, ShouldBeNil)

		executor, err = execConfig.GetExecutor(common_executors.CtrExecutorType)
		So(executor, ShouldNotBeNil)
		So(err, ShouldBeNil)

		executor, err = execConfig.GetExecutor(executors.CrosDutExecutorType)
		So(executor, ShouldNotBeNil)
		So(err, ShouldBeNil)

		executor, err = execConfig.GetExecutor(executors.CrosDutVmExecutorType)
		So(executor.GetExecutorType(), ShouldEqual, executors.CrosDutVmExecutorType)
		So(err, ShouldBeNil)

		executor, err = execConfig.GetExecutor(executors.CrosProvisionExecutorType)
		So(executor, ShouldNotBeNil)
		So(err, ShouldBeNil)

		executor, err = execConfig.GetExecutor(executors.CrosTestFinderExecutorType)
		So(executor, ShouldNotBeNil)
		So(err, ShouldBeNil)

		executor, err = execConfig.GetExecutor(executors.CacheServerExecutorType)
		So(executor, ShouldNotBeNil)
		So(err, ShouldBeNil)

		executor, err = execConfig.GetExecutor(executors.SshTunnelExecutorType)
		So(executor, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
