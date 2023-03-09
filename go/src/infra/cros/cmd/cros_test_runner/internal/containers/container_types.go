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
	CrosGcsPublishTemplatedContainerType interfaces.ContainerType = "CrosGcsPublishTemplatedContainer"
	CrosTkoPublishTemplatedContainerType interfaces.ContainerType = "CrosTkoPublishTemplatedContainer"
	CrosRdbPublishTemplatedContainerType interfaces.ContainerType = "CrosRdbPublishTemplatedContainer"
)
