// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package validation contains validation for requests.
package validation

import (
	"strings"

	"go.chromium.org/luci/common/errors"

	fleetcostAPI "infra/cros/fleetcost/api/rpc"
)

// ValidateCreateCostIndicatorRequest performs a shallow validation of cost indicator request fields.
func ValidateCreateCostIndicatorRequest(request *fleetcostAPI.CreateCostIndicatorRequest) error {
	var errs []error
	indicator := request.GetCostIndicator()
	if indicator.GetName() != "" {
		errs = append(errs, errors.New("indicator name must be empty"))
	}
	if indicator.GetBoard() == "" {
		errs = append(errs, errors.New("board cannot be empty"))
	}
	if !IsServoBoard(indicator.GetBoard()) && indicator.GetModel() == "" {
		errs = append(errs, errors.New("model cannot be empty"))
	}
	if indicator.GetCost() == nil {
		errs = append(errs, errors.New("cost must be provided"))
	}
	if indicator.GetLocation().Number() != 0 {
		errs = append(errs, errors.New("must provide valid location"))
	}
	return errors.Append(errs...)
}

// IsServoBoard checks if the board is a servo board.
func IsServoBoard(board string) bool {
	return strings.Contains(board, "servo")
}
