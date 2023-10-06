// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"testing"

	"infra/cros/cmd/common_lib/common_configs"
	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/build/api"
)

func TestGetContainer_UnsupportedContainerType(t *testing.T) {
	t.Parallel()
	Convey("Unsupported container type", t, func() {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := common_configs.NewContainerConfig(ctr, nil, false)
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
		contConfig := common_configs.NewContainerConfig(ctr, getMockContainerImagesInfo(), false)

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
		So(container.GetContainerType(), ShouldEqual, containers.CacheServerTemplatedContainerType)
		So(err, ShouldBeNil)
	})
}

func getMockContainerImagesInfo() map[string]*api.ContainerImageInfo {
	return map[string]*api.ContainerImageInfo{
		"cros-dut":          getMockedContainerImageInfo(),
		"cros-provision":    getMockedContainerImageInfo(),
		"cros-test":         getMockedContainerImageInfo(),
		"cros-publish":      getMockedContainerImageInfo(),
		"cros-test-finder":  getMockedContainerImageInfo(),
		"cache-server":      getMockedContainerImageInfo(),
		"vm-provision":      getMockedContainerImageInfo(),
		"android-provision": getMockedContainerImageInfo(),
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
