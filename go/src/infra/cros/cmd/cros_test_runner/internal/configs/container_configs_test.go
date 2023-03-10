// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"infra/cros/cmd/cros_test_runner/internal/containers"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/build/api"
)

func TestGetContainer_UnsupportedContainerType(t *testing.T) {
	t.Parallel()
	Convey("Unsupported container type", t, func() {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewCftContainerConfig(ctr, nil)
		container, err := contConfig.GetContainer(containers.UnsupportedContainerType)
		So(container, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestGetContainer_SupportedContainerType(t *testing.T) {
	t.Parallel()
	Convey("Supported container type", t, func() {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewCftContainerConfig(ctr, getMockContainerImagesInfo())

		container, err := contConfig.GetContainer(containers.CrosDutTemplatedContainerType)
		So(container, ShouldNotBeNil)
		So(err, ShouldBeNil)

		container, err = contConfig.GetContainer(containers.CrosProvisionTemplatedContainerType)
		So(container, ShouldNotBeNil)
		So(err, ShouldBeNil)

		container, err = contConfig.GetContainer(containers.CrosTestFinderTemplatedContainerType)
		So(container, ShouldNotBeNil)
		So(err, ShouldBeNil)

		container, err = contConfig.GetContainer(containers.CacheServerTemplatedContainerType)
		So(container, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func getMockContainerImagesInfo() map[string]*api.ContainerImageInfo {
	return map[string]*api.ContainerImageInfo{
		"cros-dut":         getMockedContainerImageInfo(),
		"cros-provision":   getMockedContainerImageInfo(),
		"cros-test":        getMockedContainerImageInfo(),
		"cros-publish":     getMockedContainerImageInfo(),
		"cros-test-finder": getMockedContainerImageInfo(),
		"cache-server":     getMockedContainerImageInfo(),
	}
}

func getMockedContainerImageInfo() *api.ContainerImageInfo {
	return &api.ContainerImageInfo{
		Name:   "name",
		Digest: "digest",
		Tags:   []string{"tag1"},
		Repository: &api.GcrRepository{
			Hostname: "hostName",
			Project:  "project",
		},
	}
}
