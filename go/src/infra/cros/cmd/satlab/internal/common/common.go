// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"fmt"
	"strings"
)

var IgnoreInternalFlags = map[string]bool{
	"satlab-id": false,
	"skip-dns":  false,
}

func WithInternalFlags(o ...map[string]bool) map[string]bool {
	m := make(map[string]bool)
	for k, v := range IgnoreInternalFlags {
		m[k] = v
	}
	for _, additional := range o {
		for k, v := range additional {
			m[k] = v
		}
	}
	return m
}

func MaybePrepend(satlabPrefix string, content string) string {
	if strings.HasPrefix(content, satlabPrefix) {
		return content
	}
	return fmt.Sprintf("%s-%s", satlabPrefix, content)
}
