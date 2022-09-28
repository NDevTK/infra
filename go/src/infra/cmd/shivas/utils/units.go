// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// separators is a regex that captures thousands unit separators
var separators = regexp.MustCompile(`[\., ]`)

var units = map[string]int64{
	"B":   1,
	"KB":  (1000),
	"MB":  (1000 * 1000),
	"GB":  (1000 * 1000 * 1000),
	"KiB": 1 << (10 * 1),
	"MiB": 1 << (10 * 2),
	"GiB": 1 << (10 * 3),
}

// TrimByteString removes whitespace and numeric formatting
func TrimByteString(s string) string {
	s = strings.TrimSpace(s)
	return separators.ReplaceAllLiteralString(s, "")
}

// ConvertToBytes takes a byte string and converts it to its integer form
//
// eg. '5KB' -> 5000
func ConvertToBytes(s string) (num int64, err error) {
	s = TrimByteString(s)
	if len(s) == 0 {
		return 0, nil
	}

	i := strings.LastIndexAny(s, "0123456789") + 1
	unit := s[i:]
	num, err = strconv.ParseInt(s[:i], 10, 64)
	if err != nil {
		return 0, err
	}

	if _, present := units[unit]; present {
		num *= units[unit]
		return num, nil
	} else if len(unit) == 0 {
		return num, nil
	} else {
		return 0, errors.New("unrecognized unit suffix")
	}
}
