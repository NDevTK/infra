// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package iputil

import (
	"fmt"
	"net"
)

// IsCanonicalIP returns true if the underlying IP object is 16 bytes long.
//
// Internally, net.IP can be 4 bytes or it can be 16 bytes long.
// IPv4 addresses also have a 16 byte representation with a specific prefix.
func IsCanonicalIP(ip net.IP) bool {
	return len(ip) == 16
}

// MustParseIP parses an IP address and panics if it's invalid.
func MustParseIP(x string) net.IP {
	ip := net.ParseIP(x)
	if ip == nil {
		panic(fmt.Sprintf("invalid ip address: %q", x))
	}
	return ip
}

// incrByte takes a byte, increments it, and returns a boolean indicating whether it overflowed or not.
func incrByte(x byte) (res byte, overflow bool) {
	return x + 1, x == 255
}

// RawIncr takes an IP address and increments it in an abstraction-breaking way. It doesn't respect submasks, for example.
func RawIncr(ip net.IP) (res net.IP, overflow bool) {
	overflow = true
	if len(ip) == 0 {
		return
	}
	res = make([]byte, len(ip))
	if n := copy(res, ip); n != len(ip) {
		panic("internal error in ../util/iputil/iptuil.go")
	}
	for i := -1 + len(ip); i >= 0; i-- {
		if !overflow {
			break
		}
		res[i], overflow = incrByte(ip[i])
	}
	return
}
