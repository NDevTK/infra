// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package constants

type DefaultLeaseParams struct {
	// Default disk size to use for VM creation
	DefaultDiskSize int64
	// Default machine type to use for VM creation
	DefaultMachineType string
	// Default network to use for VM creation
	DefaultNetwork string
	// Default GCP Project to use
	DefaultProject string
	// Default region (zone) to use
	DefaultRegion string
	// Default duration of lease (in minutes)
	DefaultLeaseDuration int64
}

var (
	// Default VM Leaser parameters for dev service
	DevDefaultParams = DefaultLeaseParams{
		// Default disk size to use for VM creation
		DefaultDiskSize: 20,
		// Default machine type to use for VM creation
		DefaultMachineType: "e2-medium",
		// Default network to use for VM creation
		DefaultNetwork: "global/networks/default",
		// Default GCP Project to use
		DefaultProject: "chrome-fleet-vm-leaser-dev",
		// Default region (zone) to use
		DefaultRegion: "us-central1-a",
		// Default duration of lease (in minutes)
		DefaultLeaseDuration: 600,
	}

	// Default VM Leaser parameters for prod service
	ProdDefaultParams = DefaultLeaseParams{
		// Default disk size to use for VM creation
		DefaultDiskSize: 20,
		// Default machine type to use for VM creation
		DefaultMachineType: "n2-standard-4",
		// Default network to use for VM creation
		DefaultNetwork: "global/networks/default",
		// Default GCP Project to use
		DefaultProject: "chrome-fleet-vm-leaser-prod",
		// Default region (zone) to use
		DefaultRegion: "us-central1-a",
		// Default duration of lease (in minutes)
		DefaultLeaseDuration: 600,
	}
)
