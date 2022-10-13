// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package identifiers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/karte/internal/lex64"
)

// Further fields may be added after Disambiguation without breaking backward-compatibility.
// However, adding a new field before a field that currently exists WILL break backward compatibility.
// If you are going to do this, please change the version.
type idInfo struct {
	// Should be "zzzz" initially.
	Version        string
	CoarseTime     uint64
	FineTime       uint32
	Disambiguation uint32
}

// VersionlessBytes converts an IDInfo into bytes. Note that we use big-endian order so that lexicographical comparisons of IDInfo
// correspond to lexicographical byte comparisons.
func (i *idInfo) VersionlessBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	for _, x := range []interface{}{
		i.CoarseTime,
		i.FineTime,
		i.Disambiguation,
	} {
		if err := binary.Write(buf, binary.BigEndian, x); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// Encoded converts an idInfo into lex64, which preserves lexicographic order.
func (i *iDInfo) Encoded() (string, error) {
	if i.Version == "" {
		return "", errors.Reason("encoded: version cannot be empty").Err()
	}
	bytes, err := i.VersionlessBytes()
	if err != nil {
		return "", errors.Annotate(err, "encoded").Err()
	}
	encoded, err := lex64.Encode(bytes, false)
	if err != nil {
		return "", errors.Annotate(err, "encoded").Err()
	}
	return fmt.Sprintf("%s%s", i.Version, encoded), nil
}

// Time returns the time component associated with IDInfo.
func (i *idInfo) Time() time.Time {
	return time.Unix(int64(i.CoarseTime), int64(i.FineTime)).UTC()
}

// The first four bytes of a Karte action identifier are the version.
const VersionPrefixLength = 4

// GetIDVersion gets the id version from a serialized ID.
func GetIDVersion(serialized string) string {
	if len(serialized) >= VersionPrefixLength {
		return serialized[:VersionPrefixLength]
	}
	return ""
}
