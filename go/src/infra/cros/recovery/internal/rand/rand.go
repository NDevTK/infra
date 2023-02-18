// Copyright 2023 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rand

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// String generates randomg string with expected length.
func String(length int) string {
	return stringWithCharset(length, charset)
}

// stringWithCharset generates string based on charset and required length.
func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
