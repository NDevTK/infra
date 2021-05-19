// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package swmbot

import (
	"encoding/json"

	"infra/cros/dutstate"
)

// LocalDUTState contains persistent DUT information that is cached on the
// Skylab drone.
type LocalDUTState struct {
	HostState               dutstate.State          `json:"-"`
	ProvisionableLabels     ProvisionableLabels     `json:"provisionable_labels"`
	ProvisionableAttributes ProvisionableAttributes `json:"provisionable_attributes"`
}

// ProvisionableLabels stores provisionable labels for a DUT.
type ProvisionableLabels map[string]string

// ProvisionableAttributes stores provisionable attributes for a DUT.
type ProvisionableAttributes map[string]string

// Marshal returns the encoding of the BotInfo.
func Marshal(bi *LocalDUTState) ([]byte, error) {
	return json.Marshal(bi)
}

// Unmarshal decodes BotInfo from the encoded data.
func Unmarshal(data []byte, bi *LocalDUTState) error {
	if err := json.Unmarshal(data, bi); err != nil {
		return err
	}
	if bi.ProvisionableLabels == nil {
		bi.ProvisionableLabels = make(map[string]string)
	}
	if bi.ProvisionableAttributes == nil {
		bi.ProvisionableAttributes = make(map[string]string)
	}
	return nil
}
