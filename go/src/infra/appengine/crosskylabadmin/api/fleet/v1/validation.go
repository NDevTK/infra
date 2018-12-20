// Copyright 2018 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fleet

import (
	"errors"
)

// Validate returns an error if r is invalid.
func (r *TriggerRepairOnIdleRequest) Validate() error {
	if r.IdleDuration == nil {
		return errors.New("idleDuration is required")
	}
	if r.Priority == 0 {
		return errors.New("must specify priority greater than 0")
	}
	return nil
}

// Validate returns an error if r is invalid.
func (r *TriggerRepairOnRepairFailedRequest) Validate() error {
	if r.TimeSinceLastRepair == nil {
		return errors.New("lastRepairDuration is required")
	}
	if r.Priority == 0 {
		return errors.New("must specify priority greater than 0")
	}
	return nil
}

// Validate returns an error if r is invalid.
func (r *EnsurePoolHealthyRequest) Validate() error {
	if r.DutSelector == nil {
		return errors.New("must set dut_selector")
	}
	if r.SparePool == "" {
		return errors.New("must set spare_pool")
	}
	if r.TargetPool == "" {
		return errors.New("must set target_pool")
	}
	return nil
}

// Validate returns an error if r is invalid.
func (r *ResizePoolRequest) Validate() error {
	if r.DutSelector == nil {
		return errors.New("must set dut_selector")
	}
	if r.SparePool == "" {
		return errors.New("must set spare_pool")
	}
	if r.TargetPool == "" {
		return errors.New("must set target_pool")
	}
	return nil
}

// Validate returns an error if r is invalid.
func (r *DeactivateDutRequest) Validate() error {
	if r.DutId == "" {
		return errors.New("must set dut_id")
	}
	return nil
}

// Validate returns an error if r is invalid.
func (r *ActivateDutRequest) Validate() error {
	if r.DutId == "" {
		return errors.New("must set dut_id")
	}
	return nil
}
