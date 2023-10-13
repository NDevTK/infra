// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package build

import (
	"context"
	"errors"
	"fmt"
	"time"

	"infra/build/siso/o11y/clog"
	"infra/build/siso/o11y/trace"
	"infra/build/siso/reapi"
)

func (b *Builder) runRemote(ctx context.Context, step *Step) error {
	if fastStep, ok := fastDepsCmd(ctx, b, step); ok {
		ok, err := b.tryFastStep(ctx, step, fastStep)
		if ok {
			return err
		}
	}
	step.setPhase(stepPreproc)
	b.preprocSema.Do(ctx, func(ctx context.Context) error {
		preprocCmd(ctx, b, step)
		return nil
	})
	dedupInputs(ctx, step.cmd)
	err := b.runRemoteStep(ctx, step)
	if err == nil {
		// need to check remote outptus matches cmd.Outputs?
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return err
	}
	if errors.Is(err, reapi.ErrBadPlatformContainerImage) {
		return err
	}
	if errors.Is(err, errNotRelocatable) {
		clog.Errorf(ctx, "not relocatable: %v", err)
		return err
	}
	if experiments.Enabled("no-fallback", "remote-exec %s failed. no-fallback", step) {
		return fmt.Errorf("remote-exec %s failed no-fallback: %w", step.cmd.ActionDigest(), err)
	}
	step.metrics.IsRemote = false
	step.metrics.Fallback = true
	msgs := cmdOutput(ctx, "FALLBACK", step.cmd, step.def.Binding("command"), step.def.RuleName(), err)
	b.logOutput(ctx, msgs, false)
	err = b.execLocal(ctx, step)
	if err != nil {
		return err
	}
	return b.outputs(ctx, step)
}

func (b *Builder) tryFastStep(ctx context.Context, step, fastStep *Step) (bool, error) {
	// allow local run if remote exec is not set.
	// i.e. don't run local fallback due to remote exec failure
	// because it might be bad fast-deps.
	fctx, fastSpan := trace.NewSpan(ctx, "fast-deps-run")
	err := b.runRemoteStep(fctx, fastStep)
	fastSpan.Close(nil)
	if err == nil {
		step.metrics = fastStep.metrics
		step.metrics.DepsLog = true
		msgs := cmdOutput(ctx, "SUCCESS:", fastStep.cmd, step.def.Binding("command"), step.def.RuleName(), nil)
		clog.Infof(ctx, "fast done err=%v", err)
		if len(msgs) > 0 {
			b.logOutput(ctx, msgs, step.cmd.Console)
			if experiments.Enabled("fail-on-stdouterr", "step %s emit stdout/stderr", step) {
				return true, fmt.Errorf("%s emit stdout/stderr", step)
			}
		}
		return true, b.done(ctx, fastStep)
	}
	if errors.Is(err, context.Canceled) {
		return true, err
	}
	if errors.Is(err, reapi.ErrBadPlatformContainerImage) {
		// RBE returns permission denied when
		// platform container image are not available
		// on RBE worker.
		msgs := cmdOutput(ctx, "FAILED[badContainer]:", fastStep.cmd, step.def.Binding("command"), fastStep.def.RuleName(), err)
		b.logOutput(ctx, msgs, step.cmd.Console)
		return true, err
	}
	step.metrics.DepsLogErr = true
	if experiments.Enabled("no-fast-deps-fallback", "fast-deps %s failed", step) {
		return true, fmt.Errorf("fast-deps failed: %w", err)
	}
	return false, nil
}

func (b *Builder) runRemoteStep(ctx context.Context, step *Step) error {
	if b.cache != nil && b.reCacheEnableRead {
		err := b.cacheSema.Do(ctx, func(ctx context.Context) error {
			start := time.Now()
			step.metrics.ActionStartTime = IntervalMetric(start.Sub(b.start))
			err := b.cache.GetActionResult(ctx, step.cmd)
			if err != nil {
				return err
			}
			err = b.outputs(ctx, step)
			if err != nil {
				return err
			}
			b.progressStepCacheHit(ctx, step)
			step.metrics.RunTime = IntervalMetric(time.Since(start))
			step.metrics.done(ctx, step)
			step.metrics.Cached = true
			// need to update deps for cache hit for deps=gcc, msvc.
			// even if cache hit, deps should be updated with gcc depsfile,
			// or with msvc showIncludes outputs.
			err = b.updateDeps(ctx, step)
			if err != nil {
				clog.Warningf(ctx, "failed to update deps %s: %v", step, err)
			}
			return nil
		})
		if err == nil {
			return nil
		}
		clog.Infof(ctx, "cmd cache miss: %v", err)
	}
	err := b.execRemote(ctx, step)
	if err != nil {
		return err
	}
	return b.outputs(ctx, step)
}
