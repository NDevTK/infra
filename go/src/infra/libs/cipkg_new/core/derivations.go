// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package core

import (
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"hash"
	"strings"
)

var (
	ErrUnknownDerivationKind = errors.New("unknown derivation kind")
)

func GetDerivationID(drv *Derivation) (string, error) {
	h := sha256.New()
	if err := hashDerivation(h, drv); err != nil {
		return "", err
	}

	// We want to keep the hash as short as possible to avoid reaching the path
	// length limit on windows.
	// Using base32 instead of base64 because filesystem is not promised to be
	// case-sensitive.
	enc := base32.HexEncoding.WithPadding(base32.NoPadding)
	return strings.ToLower(enc.EncodeToString(h.Sum(nil)[:16])), nil
}

func hashDerivation(h hash.Hash, drv *Derivation) error {
	if len(drv.FixedOutput) != 0 {
		_, err := h.Write([]byte(drv.FixedOutput))
		return err
	}
	return StableHash(h, drv)
}
