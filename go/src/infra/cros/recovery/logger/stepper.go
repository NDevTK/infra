// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package logger

import (
	"context"
)

// Stepper represents a simple interface for reporting steps.
type Stepper interface {
	// StartStep starts new step with provided name.
	StartStep(ctx context.Context, name string) Step
}

// Step represents a single step to track
type Step interface {
	Close(ctx context.Context, err error)
}

// Creates default stepper if not provided by recovery clients.
func NewStepper(log Logger) Stepper {
	return &simpleStepper{
		log: log,
	}
}

// simpleStepper simple representation of Stepper interface.
type simpleStepper struct {
	// log to print when step started and stopped.
	log Logger
}

// StartStep creates and start steps
func (s *simpleStepper) StartStep(ctx context.Context, name string) Step {
	step := &simpleStep{
		name: name,
		log:  s.log,
	}
	s.log.Info("Step %s: started", step.name)
	return step
}

// simpleStep simple representation of Step interface.
type simpleStep struct {
	name string
	log  Logger
}

// Close logs a step closing action.
func (s *simpleStep) Close(ctx context.Context, err error) {
	if err != nil {
		s.log.Info("Step %s: closed with error: %S", s.name, err)
	} else {
		s.log.Info("Step %s: closed", s.name)
	}
}
