// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package execs

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.chromium.org/chromiumos/config/go/api/test/xmlrpc"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/types/known/durationpb"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// Local implementation of components.Servod.
type iServod struct {
	dut *tlw.Dut
	a   tlw.Access
}

// NewServod returns a struct of type components.Servod that allowes communication with servod service.
func (ei *ExecInfo) NewServod() components.Servod {
	return &iServod{
		dut: ei.GetDut(),
		a:   ei.GetAccess(),
	}
}

// Call calls servod method with params.
func (s *iServod) Call(ctx context.Context, method string, timeout time.Duration, args ...interface{}) (*xmlrpc.Value, error) {
	log.Debugf(ctx, "Servod call %q with %v: starting...", method, args)
	res := s.a.CallServod(ctx, &tlw.CallServodRequest{
		Resource: s.dut.Name,
		Method:   method,
		Args:     packToXMLRPCValues(args...),
		Timeout:  durationpb.New(timeout),
	})
	if res.Fault {
		return nil, errors.Reason("call %q: %s", method, res.GetValue().GetScalarOneof()).Err()
	}
	log.Debugf(ctx, "Servod call %q with %v: received %#v", method, args, res.GetValue().GetScalarOneof())
	return res.Value, nil
}

// Get read value by requested command.
func (s *iServod) Get(ctx context.Context, command string) (*xmlrpc.Value, error) {
	if command == "" {
		return nil, errors.Reason("get: command is empty").Err()
	}
	v, err := s.Call(ctx, components.ServodGet, components.ServodDefaultTimeout, command)
	return v, errors.Annotate(err, "get %q", command).Err()
}

// Set sets value to provided command.
func (s *iServod) Set(ctx context.Context, command string, val interface{}) error {
	if command == "" {
		return errors.Reason("set: command is empty").Err()
	}
	if val == nil {
		return errors.Reason("set %q: value is empty", command).Err()
	}
	_, err := s.Call(ctx, components.ServodSet, components.ServodDefaultTimeout, command, val)
	return errors.Annotate(err, "set %q with %v", command, val).Err()
}

// Has verifies that command is known.
// Error is returned if the control is not listed in the doc.
func (s *iServod) Has(ctx context.Context, command string) error {
	if command == "" {
		return errors.Reason("has: command not specified").Err()
	}
	_, err := s.Call(ctx, components.ServodDoc, components.ServodDefaultTimeout, command)
	return errors.Annotate(err, "has: %q is not know", command).Err()
}

// Port provides port used for running servod daemon.
func (s *iServod) Port() int {
	return int(s.dut.GetChromeos().GetServo().GetServodPort())
}

// packToXMLRPCValues packs values to XMLRPC structs.
func packToXMLRPCValues(values ...interface{}) []*xmlrpc.Value {
	var r []*xmlrpc.Value
	for _, val := range values {
		if val == nil {
			continue
		}
		switch v := val.(type) {
		case string:
			r = append(r, &xmlrpc.Value{
				ScalarOneof: &xmlrpc.Value_String_{
					String_: v,
				},
			})
		case bool:
			r = append(r, &xmlrpc.Value{
				ScalarOneof: &xmlrpc.Value_Boolean{
					Boolean: v,
				},
			})
		case int:
			r = append(r, &xmlrpc.Value{
				ScalarOneof: &xmlrpc.Value_Int{
					Int: int32(v),
				},
			})
		case float64:
			r = append(r, &xmlrpc.Value{
				ScalarOneof: &xmlrpc.Value_Double{
					Double: v,
				},
			})
		default:
			// TODO(otabek@): Extend for more type if required. For now recovery is not using these types.
			message := fmt.Sprintf("%q is not a supported yet to be pack XMLRPC Value ", reflect.TypeOf(val))
			r = append(r, &xmlrpc.Value{
				ScalarOneof: &xmlrpc.Value_String_{
					String_: message,
				},
			})
		}
	}
	return r
}
