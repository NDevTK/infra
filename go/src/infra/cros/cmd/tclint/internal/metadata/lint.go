// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package metadata provides functions to lint Chrome OS integration test
// metadata.
package metadata

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"go.chromium.org/chromiumos/config/go/api/test/metadata/v1"
	"go.chromium.org/luci/common/errors"
)

// Result contains diagnostic messages from metadata lint.
type Result struct {
	Errors errors.MultiError
}

// Merge merges another result into the current result.
func (r *Result) Merge(o Result) {
	r.Errors = append(r.Errors, o.Errors...)
}

// MergeWithContext merges another result into the current result, prefixed with
// some context.
func (r *Result) MergeWithContext(o Result, fmt string, args ...interface{}) {
	for _, err := range o.Errors {
		// This captures the wrong stack frame. errors.Annotate() doesn't have
		// a way to specify skipping N frames (similar to testing.T.Helper())
		// yet. We don't actually render the stack trace, so this is OK.
		r.Errors = append(r.Errors, errors.Annotate(err, fmt, args...).Err())
	}
}

// AppendError appends an error to result.
func (r *Result) AppendError(fmt string, args ...interface{}) {
	r.Errors = append(r.Errors, errors.Reason(fmt, args...).Err())
}

// AppendErrorWithContext appends an error to result, prefixed with some
// context.
func (r *Result) AppendErrorWithContext(err error, fmt string, args ...interface{}) {
	r.Errors = append(r.Errors, errors.Annotate(err, fmt, args...).Err())
}

// Display returns a user-friendly display of diagnostics from a Result.
//
// Unlike String(), the result of Display() is not intended to be embedded in
// single-line context.
func (r *Result) Display() string {
	ss := []string{}
	for _, err := range r.Errors {
		ss = append(ss, fmt.Sprintf("error: %s", err.Error()))
	}
	return strings.Join(ss, "\n")
}

func errorResult(fmt string, args ...interface{}) Result {
	return Result{
		Errors: errors.NewMultiError(errors.Reason(fmt, args...).Err()),
	}
}

// Lint checks a given metadata specification for violations of requirements
// stated in the API definition.
func Lint(spec *metadata.Specification) Result {
	if len(spec.GetRemoteTestDrivers()) == 0 {
		return errorResult("Specification must contain non-zero RemoteTestDriver")
	}

	result := Result{}
	for _, rtd := range spec.RemoteTestDrivers {
		result.Merge(lintRemoteTestDriver(rtd))
	}
	return result
}

func lintRemoteTestDriver(rtd *metadata.RemoteTestDriver) Result {
	return lintRemoteTestDriverName(rtd.GetName())
}

const (
	remoteTestDriverCollection = "remoteTestDrivers"
)

func lintRemoteTestDriverName(n string) Result {
	result := Result{}
	tag := fmt.Sprintf("RemoteTestDriver '%s'", n)
	if result.MergeWithContext(lintResourceName(n), tag); result.Errors != nil {
		return result
	}
	parts := strings.Split(n, "/")
	fmt.Println(parts)
	if len(parts) != 2 || parts[0] != remoteTestDriverCollection {
		result.AppendError("%s: name must be of the form '%s/{remoteTestDriver}'", tag, remoteTestDriverCollection)
	}
	return result
}

// lintResourceName lints resource names.
//
// This lint enforces some rules in addition to the  recommendations in
// https://aip.dev/122.
//
// The returned results _do not_ add the argument as a context in diagnostic
// messages, because the caller can provide better context about the object
// being named (e.g. "RemoteTest Driver <name>" instead of "<name>").
func lintResourceName(n string) Result {
	if n == "" {
		return errorResult("name must be non-empty (https://aip.dev/122)")
	}

	result := Result{}
	u, err := url.Parse(n)
	if err != nil {
		result.AppendErrorWithContext(err, "parse name")
		return result
	}

	if u.Scheme != "" {
		result.AppendError("name must be a URL path component (https://aip.dev/122), found non-empty scheme %s", u.Scheme)
	}
	if u.Opaque != "" {
		result.AppendError("name must be a URL path component (https://aip.dev/122), found non-empty opaque data %s", u.Opaque)
	}
	if u.User != nil {
		result.AppendError("name must be a URL path component (https://aip.dev/122), found non-empty user information %s", u.User.String())
	}
	if u.Host != "" {
		result.AppendError("name must be a URL path component (https://aip.dev/122), found non-empty host %s", u.Host)
	}
	if u.Fragment != "" {
		result.AppendError("resource versions are not yet supported, found version %s", u.Fragment)
	}

	if u.Path == "" {
		result.AppendError("name must be a non-empty URL path component (https://aip.dev/122), found empty path")
		return result
	}

	if strings.HasPrefix(u.Path, "/") {
		result.AppendError("name must be a URL relative path component (https://aip.dev/122), found absolute path %s", u.Path)
	}
	if strings.HasSuffix(u.Path, "/") {
		result.AppendError("name must not contain a trailing '/' (https://aip.dev/122), found trailing '/' in %s", u.Path)
	}
	if !isASCII(u.Path) {
		result.AppendError("name must only use ASCII characters, found non-ASCII chracters in %s", u.Path)
	}
	return result
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
