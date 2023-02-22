// Copyright 2023 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package scopes

// contextKeyType is a unique type for context keys.
type contextKeyType string

const (
	ctxParamsKey             contextKeyType = "ctx_params_key"
	ctxConfigurationScopeKey contextKeyType = "ctx_configuration_scope_key"
)
