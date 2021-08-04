// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"fmt"
	"strings"
)

// DefaultZone is the default value for the zone command line flag.
const DefaultZone = "satlab"

// MaybePrepend adds a prefix with a leading dash unless the string already
// begins with the prefix in question.
func MaybePrepend(satlabPrefix string, content string) string {
	if strings.HasPrefix(content, satlabPrefix) {
		return content
	}
	return fmt.Sprintf("%s-%s", satlabPrefix, content)
}
