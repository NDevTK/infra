// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"errors"
	"fmt"
)

// ValidateVmLeaserBackend validates inputs of CreateVmInstanceRequest to be
// used with VmLeaserBackend.
func (r *CreateVmInstanceRequest) ValidateVmLeaserBackend() error {
	if r.GetConfig() == nil {
		return errors.New("invalid argument: no config found")
	}
	vmLeaserBackend := r.GetConfig().GetVmLeaserBackend()
	if vmLeaserBackend == nil {
		return fmt.Errorf("invalid argument: bad backend: want vmleaser, got %v", r.GetConfig())
	}
	if err := vmLeaserBackend.GetVmRequirements().Validate(); err != nil {
		return fmt.Errorf("invalid config argument: %w", err)
	}
	return nil
}
