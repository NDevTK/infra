// Copyright 2023 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rand

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	t.Parallel()
	for c := 0; c < 50; c++ {
		l := c
		name := fmt.Sprintf("case %d", l)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			r := String(l)
			if len(r) != l {
				t.Errorf("%q -> length does not match", name)
			}
		})
	}
}
