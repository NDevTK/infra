// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package flagx

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
)

// mapVar implements the Value interface and allows the map to be modified.
type mapVar struct {
	handle *map[string]string
}

// MapToFlagValue takes an initial map and produces a flag variable that can be
// set from command line arguments
func MapToFlagValue(m *map[string]string) flag.Value {
	if m == nil {
		panic("Argument to Dims must be non-nil pointer to map!")
	}
	return mapVar{handle: m}
}

// String returns the default value for dimensions represented as a string.
// The default value is an empty map, which stringifies to an empty string.
func (mapVar) String() string {
	return ""
}

// Set populates the dims map with comma-delimited key-value pairs.
func (mv mapVar) Set(newval string) error {
	if mv.handle == nil {
		panic("MapVar handle must be pointing at a map[string]string!")
	}
	if *mv.handle == nil {
		*mv.handle = make(map[string]string)
	}
	// strings.Split, if given an empty string, will produce a
	// slice containing a single string.
	if newval == "" {
		return nil
	}
	m := *mv.handle
	for _, entry := range strings.Split(newval, ",") {
		key, val, err := splitKeyVal(entry)
		if err != nil {
			return err
		}
		if _, exists := m[key]; exists {
			return fmt.Errorf("key %q is already specified", key)
		}
		m[key] = val
	}
	return nil
}

// splitKeyVal splits a string with "=" or ":" into two key-value
// pairs, and returns an error if this is impossible.
// Strings with multiple "=" or ":" values are considered malformed.
func splitKeyVal(s string) (string, string, error) {
	re := regexp.MustCompile("[=:]")
	res := re.Split(s, -1)
	switch len(res) {
	case 2:
		return res[0], res[1], nil
	default:
		return "", "", fmt.Errorf(`string %q is a malformed key-value pair`, s)
	}
}
