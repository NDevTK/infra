// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"bytes"
	"time"
)

// TTL is the time before a config expires.
const TTL = time.Hour

// LastSeenConfig is a simple cache to avoid unnecessary cron runs.
type LastSeenConfig struct {
	Digest []byte
	Exp    time.Time
}

func (l *LastSeenConfig) WasSeen(digest []byte) bool {
	return bytes.Equal(l.Digest, digest)
}

func (l *LastSeenConfig) MarkAsSeen(digest []byte) {
	l.Digest = digest
	l.Exp = time.Now().Add(TTL)
}

func (l *LastSeenConfig) IsExpired() bool {
	return time.Now().After(l.Exp)
}
