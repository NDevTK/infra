// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package execs

import (
	"context"
	"io"
	"time"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/logger/metrics"
)

// samplePassActionExec provides example to run action which always pass.
func samplePassActionExec(ctx context.Context, i *ExecInfo) error {
	return nil
}

// sampleFailActionExec provides example to run action which always fail.
func sampleFailActionExec(ctx context.Context, i *ExecInfo) error {
	return errors.Reason("failed").Err()
}

// sampleSleepExec pauses/sleeps the program for the time duration
// in seconds specified by the actionArgs.
//
// @params: actionArgs should be in the format of:
// Ex: ["sleep:x"]
func sampleSleepExec(ctx context.Context, i *ExecInfo) error {
	argsMap := i.GetActionArgs(ctx)
	// Timeout to wait for resetting the power state. Default to be 0s.
	sleepTimeout := argsMap.AsDuration(ctx, "sleep", 0, time.Second)
	if sleepTimeout <= 0*time.Second {
		return errors.Reason("sleep: provided time duration %v is less than or equal to 0s", sleepTimeout).Err()
	}
	log.Debugf(ctx, "Sample Sleep: planning to sleep %v.", sleepTimeout)
	time.Sleep(sleepTimeout)
	return nil
}

// sampleMetricsActionExec sends a record to the metrics service.
func sampleMetricsActionExec(ctx context.Context, info *ExecInfo) error {
	// TODO(gregorynisbet): Add more interesting information to the action.
	action := &metrics.Action{}
	if info.GetMetrics() != nil {
		action.StartTime = time.Now()
		// TODO(gregorynisbet): Don't ignore error here.
		info.GetMetrics().Create(ctx, action)
		// TODO(gregorynisbet): Uncomment when update lands.
		// defer func() { args.Metrics.Update(ctx, action) }()
	}
	// Test sleeping for one nanosecond. This will cause time to pass, which will be
	// reflected in the action and therefore in Karte.
	time.Sleep(time.Nanosecond)
	action.StopTime = time.Now()
	return nil
}

// sampleDemoMetricsExec creates and sends a metrics data as example.
// The Action create abservation and custom action as part of execution.
func sampleDemoMetricsExec(ctx context.Context, info *ExecInfo) error {
	// First we add additional observation to the metric of current action.
	info.AddObservation(metrics.NewStringObservation("date", time.Now().String()))
	info.AddObservation(metrics.NewFloat64Observation("float64", 25.25))

	// Second we create a custom metric.
	action := info.NewMetric("custom-kind")
	action.Observations = append(action.Observations,
		metrics.NewStringObservation("custom_date", time.Now().String()),
		metrics.NewFloat64Observation("custom_float64", 25.25),
	)
	// Test sleeping for one second.
	time.Sleep(time.Second)
	return nil
}

// sampleStepSummaryMarkdownExec sets experimental SummaryMarkdown to new step.
func sampleStepSummaryMarkdownExec(ctx context.Context, info *ExecInfo) error {
	if len(info.GetExecArgs()) == 0 {
		return nil
	}
	step, _ := build.StartStep(ctx, "Experimental step")
	defer func() { step.End(nil) }()
	for _, msg := range info.GetExecArgs() {
		step.Modify(func(v *build.StepView) {
			if v.SummaryMarkdown != "" {
				v.SummaryMarkdown += "<br/>"
			}
			v.SummaryMarkdown += msg
		})
	}
	return nil
}

// sampleStepLogExec sets experimental logs to new step.
func sampleStepLogExec(ctx context.Context, info *ExecInfo) error {
	step, _ := build.StartStep(ctx, "Experimental step")
	defer func() { step.End(nil) }()
	for k, v := range info.GetActionArgs(ctx) {
		req := step.Log(k)
		if v != "" {
			if _, err := io.WriteString(req, v); err != nil {
				log.Debugf(ctx, "Fail to log %v.", err)
			}
		}
	}
	return nil
}

func init() {
	Register("sample_pass", samplePassActionExec)
	Register("sample_fail", sampleFailActionExec)
	Register("sample_sleep", sampleSleepExec)
	Register("sample_metrics_action", sampleMetricsActionExec)
	Register("sample_metrics", sampleDemoMetricsExec)
	Register("sample_step_summary_markdown", sampleStepSummaryMarkdownExec)
	Register("sample_step_logs", sampleStepLogExec)
}
