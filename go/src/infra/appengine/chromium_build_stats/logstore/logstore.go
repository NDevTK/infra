// Copyright 2015 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package logstore provides an access to storage of ninja_log.
package logstore

import (
	"context"
	"strings"

	"google.golang.org/appengine/v2"
)

// Bucket returns url of the given obj.
func Bucket(ctx context.Context, obj string) (string, error) {
	obj = strings.TrimPrefix(obj, "/")
	if strings.HasPrefix(obj, "upload/") {
		return appengine.DefaultVersionHostname(ctx), nil
	}
	return "chrome-goma-log", nil
}
