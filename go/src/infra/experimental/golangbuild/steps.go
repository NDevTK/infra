// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/luciexe/build"
)

// cmdStepRun calls Run on the provided command and wraps it in a build step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func cmdStepRun(ctx context.Context, stepName string, cmd *exec.Cmd, infra bool) (err error) {
	step, ctx, err := cmdStartStep(ctx, stepName, cmd)
	defer func() {
		if infra {
			err = infraWrap(err) // Failure is deemed to be an infrastructure failure.
		}
		step.End(err)
	}()
	if err != nil {
		return err
	}

	// Combine output because it's annoying to pick one of stdout and stderr
	// in the UI and be wrong.
	output := step.Log("output")
	cmd.Stdout = output
	cmd.Stderr = output

	// Run the command.
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("failed to run %q: %w", stepName, err)
		err = attachLinks(err, fmt.Sprintf("%q (combined output)", stepName), output.UILink())
		return err
	}
	return nil
}

// cmdStepOutput calls Output on the provided command and wraps it in a build step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func cmdStepOutput(ctx context.Context, stepName string, cmd *exec.Cmd, infra bool) (output []byte, err error) {
	step, ctx, err := cmdStartStep(ctx, stepName, cmd)
	defer func() {
		if infra {
			err = infraWrap(err) // Failure is deemed to be an infrastructure failure.
		}
		step.End(err)
	}()
	if err != nil {
		return nil, err
	}

	// Make sure we log stderr.
	stderr := step.Log("stderr")
	cmd.Stderr = stderr

	// Run the command and capture stdout.
	output, err = cmd.Output()

	// Log stdout before we do anything else.
	stdout := step.Log("stdout")
	stdout.Write(output)

	// Check for errors.
	if err != nil {
		err = fmt.Errorf("failed to run %q: %w", stepName, err)
		err = attachLinks(err,
			fmt.Sprintf("%q (stdout)", stepName), stdout.UILink(),
			fmt.Sprintf("%q (stderr)", stepName), stderr.UILink(),
		)
		return output, err
	}
	return output, nil
}

// cmdStartStep sets up a command step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func cmdStartStep(ctx context.Context, stepName string, cmd *exec.Cmd) (*build.Step, context.Context, error) {
	step, ctx := build.StartStep(ctx, stepName)

	// Log the full command we're executing.
	//
	// Put each env var on its own line to actually make this readable.
	envs := cmd.Env
	if envs == nil {
		envs = os.Environ()
	}
	var fullCmd bytes.Buffer
	for _, env := range envs {
		fullCmd.WriteString(env)
		fullCmd.WriteString("\n")
	}
	if cmd.Dir != "" {
		fullCmd.WriteString("PWD=")
		fullCmd.WriteString(cmd.Dir)
		fullCmd.WriteString("\n")
	}
	fullCmd.WriteString(cmd.String())
	if _, err := io.Copy(step.Log("command"), &fullCmd); err != nil {
		return step, ctx, infraWrap(err)
	}
	return step, ctx, nil
}

func infraErrorf(s string, args ...any) error {
	return build.AttachStatus(fmt.Errorf(s, args...), bbpb.Status_INFRA_FAILURE, nil)
}

func infraWrap(err error) error {
	return build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
}

func endStep(step *build.Step, errp *error) {
	step.End(*errp)
}

func endInfraStep(step *build.Step, errp *error) {
	step.End(infraWrap(*errp))
}

// attachLinks attaches name/url pairs as links to the error.
//
// These can later be retrieved via extractLinks.
//
// Passing a nil error will result in no links attached, and
// will return another nil error.
//
// TODO(mknyszek): Consider producing a non-nil error anyway
// to avoid losing information and also to catch bugs where
// we attach information to a nil error. The downside of doing
// so is that we might end up accidentally producing a non-nil
// error for a nil error, causing a spurious failure.
func attachLinks(err error, links ...string) error {
	if len(links)%2 != 0 {
		panic("attachLinks requires name/URL pairs")
	}
	if err == nil {
		return err
	}
	el, ok := err.(*errLinks)
	if !ok {
		el = &errLinks{err: err}
	}
	for i := 0; i < len(links); i += 2 {
		el.links = append(el.links, link{
			name: links[i+0],
			url:  links[i+1],
		})
	}
	return el
}

// extractLinks aggregates the links of all errLinks in the
// error chain.
//
// Accepts a nil error, but returns no links.
func extractLinks(err error) []link {
	// Walk the error chain and extract links.
	e := err
	var links []link
	for e != nil {
		// Check if there are any links to unwrap.
		if el, ok := e.(*errLinks); ok {
			links = append(links, el.links...)
		}

		// Walk errors.Join errors.
		w, ok := e.(interface{ Unwrap() []error })
		if ok {
			for _, err := range w.Unwrap() {
				links = append(links, extractLinks(err)...)
			}
			break
		}

		// Otherwise, just try to unwrap.
		e = errors.Unwrap(e)
	}
	return links
}

// errLinks is an error with arbitrary links (name/URL pairs) attached.
// *errLinks implements error. *errLinks is unwrappable.
type errLinks struct {
	err   error // Must be non-nil.
	links []link
}

func (e *errLinks) Error() string {
	return e.err.Error()
}

func (e *errLinks) Unwrap() error {
	return e.err
}

// link is a hyperlink: a URL wth a name.
type link struct {
	name, url string
}

// attachTestsFailed marks the error as having failing tests.
//
// Accepts a nil error, but also returns a nil error in that case.
func attachTestsFailed(err error) error {
	if err == nil || errorTestsFailed(err) {
		return err
	}
	return &errTestsFailed{err}
}

// errorTestsFailed reports whether the error contains an errTestsFailed marker in its chain.
func errorTestsFailed(err error) bool {
	e := err
	for e != nil {
		// Check if there are any links to unwrap.
		if _, ok := e.(*errTestsFailed); ok {
			return true
		}

		// Walk errors.Join errors.
		w, ok := e.(interface{ Unwrap() []error })
		if ok {
			for _, err := range w.Unwrap() {
				if errorTestsFailed(err) {
					return true
				}
			}
			break
		}

		// Otherwise, just try to unwrap.
		e = errors.Unwrap(e)
	}
	return false
}

// errTestsFailed is an error that marks that tests failed.
// *errTestsFailed implements error. *errTestsFailed is unwrappable.
type errTestsFailed struct {
	err error // Must be non-nil
}

func (e *errTestsFailed) Error() string {
	return e.err.Error()
}

func (e *errTestsFailed) Unwrap() error {
	return e.err
}
