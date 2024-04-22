// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/cipkg/core"
	"go.chromium.org/luci/luciexe/build"
)

type SubstepFn func(ctx context.Context, root *build.Step) error

// RootStep manages lifetime of root step.
type RootStep struct {
	id      string
	substep chan SubstepFn

	err   error
	ended chan struct{}
}

// NewRootStep creates a root step for managing steps life time in luciexe.
func NewRootStep(ctx context.Context, name, id string) *RootStep {
	r := &RootStep{
		id:      id,
		substep: make(chan SubstepFn),

		ended: make(chan struct{}),
	}
	go r.runRoot(ctx, name)

	return r
}

// ID is the unique ID for the root step.
func (r *RootStep) ID() string { return r.id }

func (r *RootStep) runRoot(ctx context.Context, name string) {
	defer close(r.ended)

	s, ctx := build.StartStep(ctx, name)
	defer s.End(r.err)

	for sub := range r.substep {
		r.err = errors.Join(r.err, sub(ctx, s))
	}
}

// IsEnded returns whether the step has been ended.
func (r *RootStep) IsEnded() bool {
	select {
	case <-r.ended:
		return true
	default:
		return false
	}
}

// RunSubstep execute the SubstepFn in the root step environment with its
// context and *build.Step. Substep still needs to create its own step context
// by calling build.StartStep in SubstepFn.
func (r *RootStep) RunSubstep(ctx context.Context, sub SubstepFn) error {
	if r.IsEnded() {
		return fmt.Errorf("root step ended")
	}

	done := make(chan error)
	r.substep <- func(ctx context.Context, root *build.Step) error {
		err := sub(ctx, root)
		done <- err
		return err
	}

	// Either current context or RootStep is canceled/finished.
	select {
	case <-ctx.Done():
		return fmt.Errorf("sub step cancled")
	case err := <-done:
		return err
	}
}

func (r *RootStep) Err() error {
	return r.err
}

func (r *RootStep) End() {
	if r.IsEnded() {
		return
	}

	close(r.substep)
	<-r.ended
}

func runStepCommand(ctx context.Context, cmd *exec.Cmd) (err error) {
	s, _ := build.StartStep(ctx, fmt.Sprintf("run command: %s", cmd.Args))
	defer s.End(err)
	stepOutput := s.Log("stdout")

	if cmd.Stdout == nil {
		cmd.Stdout = stepOutput
	} else {
		cmd.Stdout = io.MultiWriter(cmd.Stdout, stepOutput)
	}

	if cmd.Stderr == nil {
		cmd.Stderr = stepOutput
	} else {
		cmd.Stderr = io.MultiWriter(cmd.Stderr, stepOutput)
	}

	fmt.Fprintf(s.Log("execution details"), "%#v\n", cmd)

	err = cmd.Run()
	return
}

type RootSteps map[string]*RootStep

// NewRootSteps creates RootSteps for managing a lookup table for root steps in
// luciexe. This can be updated by preExecFn and used by execFn to group
// derivation execution by root steps.
func NewRootSteps() RootSteps {
	return make(RootSteps)
}

// UpdateRoot sets all package's non root dependencies' root to the package's
// root recursively.
func (rs RootSteps) UpdateRoot(ctx context.Context, pkg actions.Package) (*RootStep, error) {
	return rs.update(ctx, pkg, nil)
}

func (rs RootSteps) update(ctx context.Context, pkg actions.Package, root *RootStep) (*RootStep, error) {
	if r, ok := rs[pkg.DerivationID]; ok {
		// If pkg has root other than itself.
		if r.ID() != pkg.DerivationID {
			if root == nil {
				return nil, fmt.Errorf("top level package shouldn't belong to other root: %s, from %s", pkg.DerivationID, r.ID())
			}
			if r.ID() != root.ID() {
				return nil, fmt.Errorf("package must only belong to one root: %s, from %s and %s", pkg.DerivationID, r.ID(), root.ID())
			}
		}

		return r, nil
	}

	if root == nil || isRootStep(pkg) {
		name := pkg.Action.Metadata.GetLuciexe().GetStepName()
		if name == "" {
			name = pkg.DerivationID
		}

		root = NewRootStep(ctx, name, pkg.DerivationID)
	}
	rs[pkg.DerivationID] = root

	for _, dep := range pkg.RuntimeDependencies {
		if _, err := rs.update(ctx, dep, root); err != nil {
			return nil, err
		}
	}

	for _, dep := range pkg.BuildDependencies {
		if _, err := rs.update(ctx, dep, root); err != nil {
			return nil, err
		}
	}

	return root, nil
}

// GetRoot returns the root step for the derivation id.
func (rs RootSteps) GetRoot(id string) *RootStep { return rs[id] }

// isRootStep returns whether a package is a root step in luciexe.
// Consider a package as a root package if it's either
// a root step with name, or
// importing from host, or
// embedded during build time.
// TODO(fancl): eventually all these should be organized under a real root
// package so we can simply check the step name to decide.
func isRootStep(pkg actions.Package) bool {
	if pkg.Action.Metadata.GetLuciexe().GetStepName() != "" {
		return true
	}

	// 3pp specs are imported from local files but should be considered as
	// substep.
	if strings.HasSuffix(pkg.Action.Name, "_from_spec_def") {
		return false
	}

	// TODO(fancl): from_spec_tools should be replaced by 3pp dependencies.
	if pkg.Action.Name == "from_spec_tools" {
		return true
	}

	if copy := pkg.Action.GetCopy(); copy != nil {
		for _, f := range copy.GetFiles() {
			switch f.Content.(type) {
			case *core.ActionFilesCopy_Source_Local_, *core.ActionFilesCopy_Source_Embed_, *core.ActionFilesCopy_Source_Raw:
			default:
				return false
			}
		}
		return true
	}

	return false
}
