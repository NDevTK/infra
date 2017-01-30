// Copyright 2014 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gitiles

import (
	"fmt"
)

// StatusError is a simple error type which wraps a Http StatusCode.
type StatusError int

// Member functions ////////////////////////////////////////////////////////////

// Bad returns true iff the status code is not 2XX
func (s StatusError) Bad() bool     { return s < 200 || s >= 300 }
func (s StatusError) Error() string { return fmt.Sprintf("got status code %d", s) }
