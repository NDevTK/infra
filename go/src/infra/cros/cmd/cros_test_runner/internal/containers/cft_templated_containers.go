// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package containers

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
	"infra/cros/cmd/cros_test_runner/internal/tools/crostoolrunner"
)

func NewCrosDutTemplatedContainer(
	containerImage string,
	ctr *crostoolrunner.CrosToolRunner) interfaces.ContainerInterface {

	return NewContainer(CrosDutTemplatedContainerType, "cros-dut", containerImage, ctr, true)
}

func NewCrosProvisionTemplatedContainer(
	containerImage string,
	ctr *crostoolrunner.CrosToolRunner) interfaces.ContainerInterface {

	return NewContainer(CrosProvisionTemplatedContainerType, "cros-provision", containerImage, ctr, true)
}

func NewCrosTestTemplatedContainer(
	containerImage string,
	ctr *crostoolrunner.CrosToolRunner) interfaces.ContainerInterface {

	return NewContainer(CrosTestTemplatedContainerType, "cros-test", containerImage, ctr, true)
}

func NewCrosTestFinderTemplatedContainer(
	containerImage string,
	ctr *crostoolrunner.CrosToolRunner) interfaces.ContainerInterface {

	return NewContainer(CrosTestFinderTemplatedContainerType, "cros-test-finder", containerImage, ctr, true)
}

func NewCacheServerTemplatedContainer(
	containerImage string,
	ctr *crostoolrunner.CrosToolRunner) interfaces.ContainerInterface {

	return NewContainer(CacheServerTemplatedContainerType, "cache-server", containerImage, ctr, true)
}

func NewCrosPublishTemplatedContainer(
	contType interfaces.ContainerType,
	containerImage string,
	ctr *crostoolrunner.CrosToolRunner) interfaces.ContainerInterface {

	if contType != CrosGcsPublishTemplatedContainerType && contType != CrosTkoPublishTemplatedContainerType && contType != CrosCpconPublishTemplatedContainerType && contType != CrosRdbPublishTemplatedContainerType {
		return nil
	}
	return NewContainer(contType, "cros-publish", containerImage, ctr, true)
}
