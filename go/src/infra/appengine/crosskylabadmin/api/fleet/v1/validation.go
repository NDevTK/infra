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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Validate returns an error if r is invalid.
func (r *RemoveDutsFromDronesRequest) Validate() error {
	for _, item := range r.Removals {
		if item.DutId == "" && item.DutHostname == "" {
			return errors.New("must specify one of DutId or DutHostname")
		}
		if item.DutId != "" && item.DutHostname != "" {
			return errors.New("must specify only one of DutId and DutHostname")
		}
	}
	return nil
}

// Validate returns an error if r is invalid.
func (r *AssignDutsToDronesRequest) Validate() error {
	for _, item := range r.Assignments {
		if item.DutId == "" && item.DutHostname == "" {
			return errors.New("must specify one of DutId or DutHostname")
		}
		if item.DutId != "" && item.DutHostname != "" {
			return errors.New("must specify only one of DutId and DutHostname")
		}
	}
	return nil
}

// Validate returns an error if r is invalid.
func (r *GetDutInfoRequest) Validate() error {
	if r.Id == "" && r.Hostname == "" {
		return errors.New("one of id or hostname is required")
	}
	return nil
}

// Validate returns an error if r is invalid.
func (r *DutSelector) Validate() error {
	if r.Id == "" && r.Hostname == "" && r.Model == "" {
		return errors.New("dut_selector must not be empty")
	}
	return nil
}

// Validate returns an error if r is invalid.
func (r *DeleteDutsRequest) Validate() error {
	if len(r.Hostnames) == 0 {
		return status.Errorf(codes.InvalidArgument, "must specify at least one hostname")
	}
	return nil
}
