// Copyright 2023 The Chromium Authors
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
	"infra/cros/recovery/internal/localtlw/servod"
	"infra/cros/recovery/tlw"
)

const (
	defaultTouchhostdPort = 9992
)

// CallTouchHostd executes a command on touchhostd.
func (c *tlwClient) CallTouchHostd(ctx context.Context, req *tlw.CallTouchHostdRequest) *tlw.CallTouchHostdResponse {
	// Translator to convert error to response structure.
	fail := func(err error) *tlw.CallTouchHostdResponse {
		return &tlw.CallTouchHostdResponse{
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

	callTimeout := 30 * time.Second
	if req.GetTimeout().GetSeconds() > 0 {
		callTimeout = req.GetTimeout().AsDuration()
	}

	val, err := servod.CallServod(ctx, &servod.ServodCallRequest{
		Host:        localproxy.BuildAddr(req.GetResource()),
		SSHProvider: c.sshProvider,
		Options: &tlw.ServodOptions{
			ServodPort: int32(defaultTouchhostdPort),
		},
		CallMethod:    req.GetMethod(),
		CallArguments: req.GetArgs(),
		CallTimeout:   callTimeout,
	})
	if err != nil {
		return fail(err)
	}
	return &tlw.CallTouchHostdResponse{
		Value: val,
		Fault: false,
	}
}
