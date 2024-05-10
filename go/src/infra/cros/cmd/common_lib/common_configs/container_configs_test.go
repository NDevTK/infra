// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_configs

import (
	"testing"

	"go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/luci/common/testing/ftt"
	"go.chromium.org/luci/common/testing/truth/assert"
	"go.chromium.org/luci/common/testing/truth/should"

	"infra/cros/cmd/common_lib/containers"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
)

func TestGetContainer_UnsupportedContainerType(t *testing.T) {
	t.Parallel()
	ftt.Parallel("Unsupported container type", t, func(t *ftt.Test) {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewContainerConfig(ctr, nil, false)
		container, err := contConfig.GetContainer(containers.UnsupportedContainerType)
		assert.Loosely(t, container, should.BeNil)
		assert.Loosely(t, err, should.NotBeNil)
	})
}

func TestGetContainer_SupportedContainerType(t *testing.T) {
	t.Parallel()
	ftt.Parallel("Supported container type", t, func(t *ftt.Test) {
		ctrCipd := crostoolrunner.CtrCipdInfo{Version: "prod"}
		ctr := &crostoolrunner.CrosToolRunner{CtrCipdInfo: ctrCipd}
		contConfig := NewContainerConfig(ctr, getMockContainerImagesInfo(), false)

		container, err := contConfig.GetContainer(containers.CrosDutTemplatedContainerType)
		assert.Loosely(t, container, should.NotBeNil)
		assert.Loosely(t, err, should.BeNil)

		container, err = contConfig.GetContainer(containers.CrosProvisionTemplatedContainerType)
		assert.Loosely(t, container, should.NotBeNil)
		assert.Loosely(t, err, should.BeNil)

		container, err = contConfig.GetContainer(containers.CrosTestFinderTemplatedContainerType)
		assert.Loosely(t, container, should.NotBeNil)
		assert.Loosely(t, err, should.BeNil)

		container, err = contConfig.GetContainer(containers.CacheServerTemplatedContainerType)
		assert.Loosely(t, container.GetContainerType(), should.Equal(containers.CacheServerTemplatedContainerType))
		assert.Loosely(t, err, should.BeNil)
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
