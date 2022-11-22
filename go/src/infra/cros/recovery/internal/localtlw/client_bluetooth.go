// Copyright 2022 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package localtlw provides local implementation of TLW Access.
package localtlw

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/chromiumos/config/go/api/test/xmlrpc"

	"infra/cros/recovery/internal/localtlw/localproxy"
	"infra/cros/recovery/tlw"
)

const (
	defaultBluetoothPeerServerPort = 9992
)

// CallBluetoothPeer executes a command on bluetooth-peer service.
func (c *tlwClient) CallBluetoothPeer(ctx context.Context, req *tlw.CallBluetoothPeerRequest) *tlw.CallBluetoothPeerResponse {
	// Translator to convert error to response structure.
	fail := func(err error) *tlw.CallBluetoothPeerResponse {
		return &tlw.CallBluetoothPeerResponse{
			Value: &xmlrpc.Value{
				ScalarOneof: &xmlrpc.Value_String_{
					String_: fmt.Sprintf("call servod %q: %s", req.GetResource(), err),
				},
			},
			Fault: true,
		}
	}
	// Check if the name was detected by loaded device.
	_, err := c.getDevice(ctx, req.GetResource())
	if err != nil {
		return fail(err)
	}
	s, err := c.servodPool.Get(
		localproxy.BuildAddr(req.GetResource()),
		int32(defaultBluetoothPeerServerPort),
		func() ([]string, error) { return nil, nil })
	if err != nil {
		return fail(err)
	}
	// TODO(otabek): Change bluetooth peer's CallBluetoothPeerRequest to include timeout.
	val, err := s.Call(ctx, c.sshPool, 30*time.Second, req.GetMethod(), req.GetArgs())
	if err != nil {
		return fail(err)
	}
	return &tlw.CallBluetoothPeerResponse{
		Value: val,
		Fault: false,
	}
}
