// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package containers

import "infra/cros/cmd/cros_test_runner/internal/interfaces"

// All supported container types.
const (
	// For testing purposes only
	UnsupportedContainerType             interfaces.ContainerType = "UnsupportedContainer"
	CacheServerTemplatedContainerType    interfaces.ContainerType = "CacheServerTemplatedContainer"
	CrosProvisionTemplatedContainerType  interfaces.ContainerType = "CrosProvisionTemplatedContainer"
	CrosDutTemplatedContainerType        interfaces.ContainerType = "CrosDutTemplatedContainer"
	CrosTestTemplatedContainerType       interfaces.ContainerType = "CrosTestTemplatedContainer"
	CrosTestFinderTemplatedContainerType interfaces.ContainerType = "CrosTestFinderTemplatedContainer"
	CrosGcsPublishTemplatedContainerType interfaces.ContainerType = "CrosGcsPublishTemplatedContainer"
	CrosTkoPublishTemplatedContainerType interfaces.ContainerType = "CrosTkoPublishTemplatedContainer"
	CrosRdbPublishTemplatedContainerType interfaces.ContainerType = "CrosRdbPublishTemplatedContainer"
)

// GetContainerImageKeyFromContainerType converts a ContainerType to its commonly known string representation.
func GetContainerImageKeyFromContainerType(containerType interfaces.ContainerType) string {
	switch containerType {
	case UnsupportedContainerType:
		return ""
	case CacheServerTemplatedContainerType:
		return "cache-server"
	case CrosProvisionTemplatedContainerType:
		return "cros-provision"
	case CrosDutTemplatedContainerType:
		return "cros-dut"
	case CrosTestTemplatedContainerType:
		return "cros-test"
	case CrosTestFinderTemplatedContainerType:
		return "cros-test-finder"
	case CrosGcsPublishTemplatedContainerType, CrosTkoPublishTemplatedContainerType, CrosRdbPublishTemplatedContainerType:
		return "cros-publish"
	default:
		return ""
	}
}
