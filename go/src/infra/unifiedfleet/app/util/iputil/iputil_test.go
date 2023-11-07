// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package iputil

import (
	"net"
	"testing"

	"go.chromium.org/luci/common/testing/typed"
)

func TestIncrByte(t *testing.T) {
	x, overflow := incrByte(255)
	if x != 0 {
		t.Error("255 should overflow to 0")
	}
	if !overflow {
		t.Error("255 should overflow")
	}
}

func TestRawIncr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		ip       net.IP
		want     net.IP
		overflow bool
	}{
		{
			name:     "basic increment",
			ip:       net.IPv4(127, 0, 0, 1),
			want:     net.IPv4(127, 0, 0, 2),
			overflow: false,
		},
		{
			name:     "edge of ipv4 space",
			ip:       net.IPv4(255, 255, 255, 255),
			want:     MustParseIP("::1:0:0:0"),
			overflow: false,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, overflow := RawIncr(tt.ip)

			if diff := typed.Diff(got, tt.want); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
			if diff := typed.Diff(overflow, tt.overflow); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}
