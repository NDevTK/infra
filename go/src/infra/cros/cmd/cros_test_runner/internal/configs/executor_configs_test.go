// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"infra/cros/cmd/cros_test_runner/internal/executors"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetExecutor_UnsupportedExecutorType(t *testing.T) {
	t.Parallel()
	Convey("Unsupported executor type", t, func() {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewCftContainerConfig(ctr, nil)
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
		contConfig := NewCftContainerConfig(ctr, getMockContainerImagesInfo())
		execConfig := NewExecutorConfig(ctr, contConfig)

		executor, err := execConfig.GetExecutor(executors.NoExecutorType)
		So(executor, ShouldBeNil)
		So(err, ShouldNotBeNil)

		executor, err = execConfig.GetExecutor(executors.InvServiceExecutorType)
		So(executor, ShouldNotBeNil)
		So(err, ShouldBeNil)

		executor, err = execConfig.GetExecutor(executors.CtrExecutorType)
		So(executor, ShouldNotBeNil)
		So(err, ShouldBeNil)

		executor, err = execConfig.GetExecutor(executors.CrosDutExecutorType)
		So(executor, ShouldNotBeNil)
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
