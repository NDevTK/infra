// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

// InstanceApi is the VM instance management API that all providers implement.
type InstanceApi interface {
	// Create leases a new VM instance.
	Create(*CreateVmInstanceRequest) (*VmInstance, error)
	// Delete releases an existing VM instance.
	Delete(*VmInstance) error
	// Cleanup releases existing VM instances that match the request.
	Cleanup(*CleanupVmInstancesRequest) error
}

// ImageApi is the VM image management API that all providers implement. This is
// a MVP integration that will be replaced by cloud native approach once
// relevant GCP services are available for internal use:
// go/cros-vm-cloud-native-image-sync
type ImageApi interface {
	// GetImage treats imported image as cache keyed by builderPath. On cache-miss
	// the method will try to import image. When wait is true, the method will
	// poll the image until the image is READY, or error, or timeout. When wait is
	// false, the current status of the image is returned immediately.
	// go/cros-image-importer
	GetImage(builderPath string, wait bool) (*GceImage, error)
	// ListImages returns a list of VM images in the hard-coded GCP project in
	// cloudsdk.go. Pass empty string to `filter` to return all images. To
	// filter results specify `filter` with format described in GCP API doc:
	// https://cloud.google.com/compute/docs/reference/rest/v1/images/list
	ListImages(filter string) ([]GceImage, error)
	// DeleteImage deletes a VM image with name `imageName` from the hard-coded
	// GCP project in cloudsdk.go. When wait is true, it will wait until the
	// delete operation is finished.
	DeleteImage(imageName string, wait bool) error
}
