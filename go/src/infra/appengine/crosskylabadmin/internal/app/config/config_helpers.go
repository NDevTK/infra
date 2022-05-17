// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

// PermilleData contains information on what portion of traffic to opt
// into Prod and Latest.
type PermilleData struct {
	Prod   float64
	Latest float64
}

// ComputeProdPermille computes the most applicable prod permille
// for a device.
func (x *RolloutConfig) ComputePermilleData(hostname string) (PermilleData, error) {
	return PermilleData{
		Prod:   float64(x.GetProdPermille()),
		Latest: float64(x.GetLatestPermille()),
	}, nil
}
