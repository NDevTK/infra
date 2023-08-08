// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

// ConfigDryRun is an empty config used for a dry run.
//
// This config intentionally does nothing.
func ConfigDryRun() *Configuration {
	return &Configuration{}
}
