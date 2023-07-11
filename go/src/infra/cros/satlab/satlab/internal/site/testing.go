// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package site

// MakeEnvFlagsForTesting creates env flags while allowing us to keep fields in
// EnvFlags private, so that users are forced to use GetNamespace() instead for
// real code.
func MakeEnvFlagsForTesting(ns string) EnvFlags {
	return EnvFlags{
		namespace: ns,
	}
}
