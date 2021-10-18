// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"net/url"
	"regexp"

	protov1 "github.com/golang/protobuf/proto"
	luciproto "go.chromium.org/luci/common/proto"
	"go.chromium.org/luci/config/validation"
)

// From https://source.chromium.org/chromium/infra/infra/+/main:appengine/monorail/project/project_constants.py;l=13.
var monorailProjectRE = regexp.MustCompile(`^[a-z0-9][-a-z0-9]{0,61}[a-z0-9]$`)

const maxHysteresisPercent = 1000

func validateConfig(ctx *validation.Context, cfg *Config) {
	validateMonorailHostname(ctx, cfg.MonorailHostname)
}

func validateMonorailHostname(ctx *validation.Context, hostname string) {
	ctx.Enter("monorail_hostname")
	if hostname == "" {
		ctx.Errorf("empty value is not allowed")
	} else if _, err := url.Parse(hostname); err != nil {
		ctx.Errorf("invalid hostname: %s", hostname)
	}
	ctx.Exit()
}

// validateProjectConfigRaw deserializes the project-level config message
// and passes it through the validator.
func validateProjectConfigRaw(ctx *validation.Context, content string) *ProjectConfig {
	msg := &ProjectConfig{}
	if err := luciproto.UnmarshalTextML(content, protov1.MessageV1(msg)); err != nil {
		ctx.Errorf("failed to unmarshal as text proto: %s", err)
		return nil
	}
	validateProjectConfig(ctx, msg)
	return msg
}

func validateProjectConfig(ctx *validation.Context, cfg *ProjectConfig) {
	validateMonorail(ctx, cfg.Monorail)
	validateImpactThreshold(ctx, cfg.BugFilingThreshold, "bug_filing_threshold")
	validateBigQueryTable(ctx, cfg.ClusteredFailuresTable, "clustered_failures_table")
}

func validateMonorail(ctx *validation.Context, cfg *MonorailProject) {
	ctx.Enter("monorail")
	defer ctx.Exit()

	if cfg == nil {
		ctx.Errorf("monorail must be specified")
		return
	}

	validateMonorailProject(ctx, cfg.Project)
	validateDefaultFieldValues(ctx, cfg.DefaultFieldValues)
	validateFieldID(ctx, cfg.PriorityFieldId, "priority_field_id")
	validatePriorities(ctx, cfg.Priorities)
	validatePriorityHysteresisPercent(ctx, cfg.PriorityHysteresisPercent)
}

func validateMonorailProject(ctx *validation.Context, project string) {
	ctx.Enter("project")
	if project == "" {
		ctx.Errorf("empty value is not allowed")
	} else if !monorailProjectRE.MatchString(project) {
		ctx.Errorf("project is not a valid monorail project")
	}
	ctx.Exit()
}

func validateDefaultFieldValues(ctx *validation.Context, fvs []*MonorailFieldValue) {
	ctx.Enter("default_field_values")
	for i, fv := range fvs {
		ctx.Enter("[%v]", i)
		validateFieldValue(ctx, fv)
		ctx.Exit()
	}
	ctx.Exit()
}

func validateFieldID(ctx *validation.Context, fieldID int64, fieldName string) {
	ctx.Enter(fieldName)
	if fieldID < 0 {
		ctx.Errorf("value must be non-negative")
	}
	ctx.Exit()
}

func validateFieldValue(ctx *validation.Context, fv *MonorailFieldValue) {
	validateFieldID(ctx, fv.GetFieldId(), "field_id")
	// No validation applies to field value.
}

func validatePriorities(ctx *validation.Context, ps []*MonorailPriority) {
	ctx.Enter("priorities")
	if len(ps) == 0 {
		ctx.Errorf("at least one monorail priority must be specified")
	}
	for i, p := range ps {
		ctx.Enter("[%v]", i)
		validatePriority(ctx, p)
		ctx.Exit()
	}
	ctx.Exit()
}

func validatePriority(ctx *validation.Context, p *MonorailPriority) {
	validatePriorityValue(ctx, p.Priority)
	validateImpactThreshold(ctx, p.Threshold, "threshold")
}

func validatePriorityValue(ctx *validation.Context, value string) {
	ctx.Enter("priority")
	// Although it is possible to allow the priority field to be empty, it
	// would be rather unusual for a project to set itself up this way. For
	// now, prefer to enforce priority values are non-empty as this will pick
	// likely configuration errors.
	if value == "" {
		ctx.Errorf("empty value is not allowed")
	}
	ctx.Exit()
}

func validateImpactThreshold(ctx *validation.Context, t *ImpactThreshold, fieldName string) {
	ctx.Enter(fieldName)
	defer ctx.Exit()

	if t == nil {
		ctx.Errorf("impact thresolds must be specified")
		return
	}

	validateFailureCountThresold(ctx, t.UnexpectedFailures_1D, "unexpected_failures_1d")
	validateFailureCountThresold(ctx, t.UnexpectedFailures_3D, "unexpected_failures_3d")
	validateFailureCountThresold(ctx, t.UnexpectedFailures_7D, "unexpected_failures_7d")
}

func validatePriorityHysteresisPercent(ctx *validation.Context, value int64) {
	ctx.Enter("priority_hysteresis_percent")
	if value > maxHysteresisPercent {
		ctx.Errorf("value must not exceed %v percent", maxHysteresisPercent)
	}
	if value < 0 {
		ctx.Errorf("value must not be negative")
	}
	ctx.Exit()
}

func validateFailureCountThresold(ctx *validation.Context, threshold *int64, fieldName string) {
	ctx.Enter(fieldName)
	if threshold != nil && *threshold < 0 {
		ctx.Errorf("value must be non-negative")
	}
	ctx.Exit()
}

func validateBigQueryTable(ctx *validation.Context, t *BigQueryTable, fieldName string) {
	ctx.Enter(fieldName)
	defer ctx.Exit()

	if t == nil {
		ctx.Errorf("value must be specified")
		return
	}
	if t.Project == "" {
		ctx.Errorf("project must be specified")
	}
	if t.Dataset == "" {
		ctx.Errorf("dataset must be specified")
	}
	if t.Table == "" {
		ctx.Errorf("table must be specified")
	}
}
