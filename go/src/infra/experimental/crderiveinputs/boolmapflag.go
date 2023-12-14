// Copyright (c) 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"go.chromium.org/luci/common/errors"
)

type boolmapflag map[string]bool

var _ flag.Value = (*boolmapflag)(nil)

func (b boolmapflag) String() string {
	keys := make([]string, len(b))
	for key := range b {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for idx, k := range keys {
		keys[idx] = fmt.Sprintf("%s=%t", k, b[k])
	}
	return fmt.Sprintf("[%s]", strings.Join(keys, ","))
}

func (b boolmapflag) Set(val string) error {
	key := strings.TrimSpace(val)
	if len(key) == 0 {
		return errors.New("cannot specify an empty k=v pair")
	}

	value := ""
	idx := strings.Index(key, "=")
	switch {
	case idx == -1:
	case idx == 0:
		return errors.New("cannot have a k=v pair with empty key")

	case idx > 0:
		key, value = key[:idx], key[idx+1:]
	}

	if value == "" || value == "true" {
		b[key] = true
	} else if value == "false" {
		b[key] = false
	} else {
		return fmt.Errorf("value must be '' or 'true' or 'false', got %q", value)
	}
	return nil
}
