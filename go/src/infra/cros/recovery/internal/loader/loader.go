// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package loader provides functionality to load configuration and verify it.
package loader

import (
	"context"
	"encoding/json"
	"io"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/internal/planpb"
)

// LoadConfiguration performs loading the configuration source with data validation.
// TODO(otabek@): Add data validation for loaded config.
func LoadConfiguration(ctx context.Context, r io.Reader) (*planpb.Configuration, error) {
	log.Debug(ctx, "Load configuration: started.")
	if r == nil {
		return nil, errors.Reason("load configuration: reader is not provided").Err()
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Annotate(err, "load configuration").Err()
	}
	if len(data) == 0 {
		return nil, errors.Reason("load configuration: configuration is empty").Err()
	}
	config := planpb.Configuration{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.Annotate(err, "load configuration").Err()
	}
	// TODO(otabek@): Verify only critical action can have recovery actions.
	for _, p := range config.GetPlans() {
		if err := setAndVerifyExecs(p); err != nil {
			return nil, errors.Annotate(err, "load configuration").Err()
		}
	}
	log.Debug(ctx, "Load configuration: finished successfully.")
	return &config, nil
}

// execsExist is link to the function to check if exec function is present.
// Link created to create ability to override for local testing.
var execsExist = execs.Exist

// setAndVerifyExecs sets exec-name if missing and validate whether exec is present
// in recovery-lib.
func setAndVerifyExecs(p *planpb.Plan) error {
	for an, a := range p.GetActions() {
		if a.GetExecName() == "" {
			a.ExecName = an
		}
		if !execsExist(a.GetExecName()) {
			return errors.Reason("exec %q is not exist", a.GetExecName()).Err()
		}
	}
	return nil
}
