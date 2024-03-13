// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package provider provides a generic template
// for new Bots Provider Interfaces.
package provider

import (
	"context"
)

// BPI is a generic Provider interface.
// Future Providers need to satisfy this interface.
type BPI interface {
	UpdateConfig(context.Context, []string) error
}
