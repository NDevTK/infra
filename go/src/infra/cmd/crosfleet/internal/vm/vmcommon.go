// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vm

import (
	"regexp"
)

// sanitizeForLabel replaces all unsupported characters with _ to be compatible
// with labels on GCP.
func sanitizeForLabel(str string) string {
	re := regexp.MustCompile(`[^a-z0-9-]`)
	return re.ReplaceAllString(str, "_")
}
