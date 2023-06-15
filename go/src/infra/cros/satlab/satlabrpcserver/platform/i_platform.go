// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package platform

import "context"

type HostIdentifier struct {
	ID string
}

type IPlatform interface {
	// GetHostIdentifier get a machine identifier that should be unique and stable
	GetHostIdentifier(context.Context) (string, error)
}
