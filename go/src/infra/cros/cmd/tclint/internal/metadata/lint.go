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
)

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
		result.AppendError("name must be a URL path component (https://aip.dev/122), found non-empty scheme '%s'", u.Scheme)
	}
	if u.Opaque != "" {
		result.AppendError("name must be a URL path component (https://aip.dev/122), found non-empty opaque data '%s'", u.Opaque)
	}
	if u.User != nil {
		result.AppendError("name must be a URL path component (https://aip.dev/122), found non-empty user information '%s'", u.User.String())
	}
	if u.Host != "" {
		result.AppendError("name must be a URL path component (https://aip.dev/122), found non-empty host '%s'", u.Host)
	}
	if u.Fragment != "" {
		result.AppendError("resource versions are not yet supported, found version '%s'", u.Fragment)
	}

	if u.Path == "" {
		result.AppendError("name must be a non-empty URL path component (https://aip.dev/122), found empty path")
		return result
	}

	if strings.HasPrefix(u.Path, "/") {
		result.AppendError("name must be a URL relative path component (https://aip.dev/122), found absolute path '%s'", u.Path)
	}
	if strings.HasSuffix(u.Path, "/") {
		result.AppendError("name must not contain a trailing '/' (https://aip.dev/122), found trailing '/' in '%s'", u.Path)
	}
	if !isASCII(u.Path) {
		result.AppendError("name must only use ASCII characters, found non-ASCII chracters in '%s'", u.Path)
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
