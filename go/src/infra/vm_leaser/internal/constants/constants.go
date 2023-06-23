// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package constants

// A list of default parameters for VM Leaser
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

// A list of GCP zones for VM Leaser service to manage VMs in
var (
	UsCentral1 = []string{
		"us-central1-a",
		"us-central1-b",
		"us-central1-c",
		"us-central1-f",
	}

	UsEast1 = []string{
		"us-east1-b",
		"us-east1-c",
		"us-east1-d",
	}

	UsEast4 = []string{
		"us-east4-a",
		"us-east4-b",
		"us-east4-c",
	}

	UsWest1 = []string{
		"us-west1-a",
		"us-west1-b",
		"us-west1-c",
	}
)

// Different zonal configurations for different products
//
// AllQuotaZones is the list of zones where quota for VMs has been given. All
// other zonal configurations are defined using AllQuotaZones. Each key in
// AllQuotaZones represents a main zone and is defined in the GCP region format.
var (
	AllQuotaZones = map[string][]string{
		"us-central1": UsCentral1,
		"us-east1":    UsEast1,
		"us-east4":    UsEast4,
		"us-west1":    UsWest1,
	}

	ChromeOSZones = [][]string{
		AllQuotaZones["us-central1"],
		AllQuotaZones["us-east1"],
		AllQuotaZones["us-east4"],
		AllQuotaZones["us-west1"],
	}
)
