// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servo

import (
	"context"
	"fmt"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/logger/metrics"
)

// servoCheckServodControlExec verifies that servod supports the
// control mentioned in action args. Additionally, if actionArgs
// includes the expected value, this function will verify that the
// value returned by servod for this control matches the expected
// value.
func servoCheckServodControlExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	command := argsMap.AsString(ctx, "command", "")
	if len(command) == 0 {
		return errors.Reason("servo check servod control exec: command not provided").Err()
	}
	const expectedStringKey = "expected_string_value"
	const expectedIntKey = "expected_int_value"
	const expectedIntGreaterKey = "expected_int_value_greater"
	const expectedIntLessKey = "expected_int_value_less"
	const expectedFloatKey = "expected_float_value"
	const expectedBoolKey = "expected_bool_value"
	var servodControlValue string
	defer func() {
		info.AddObservation(metrics.NewStringObservation(fmt.Sprintf("servod:%s", command), servodControlValue))
	}()
	if argsMap.Has(expectedStringKey) {
		expectedValue := argsMap.AsString(ctx, expectedStringKey, "")
		controlValue, err := servodGetString(ctx, info.NewServod(), command)
		if err != nil {
			servodControlValue = err.Error()
			return errors.Annotate(err, "servo check servod control exec").Err()
		}
		servodControlValue = controlValue
		log.Infof(ctx, "Compare (String), expected value %q, actual value %q", expectedValue, controlValue)
		if controlValue != expectedValue {
			return errors.Reason("compare (string): expected value %q, actual value %q do not match.", expectedValue, controlValue).Err()
		}
	} else if argsMap.Has(expectedIntKey) || argsMap.Has(expectedIntGreaterKey) || argsMap.Has(expectedIntLessKey) {
		controlValue, err := servodGetInt(ctx, info.NewServod(), command)
		if err != nil {
			servodControlValue = err.Error()
			return errors.Annotate(err, "servo check servod control exec").Err()
		}
		servodControlValue = fmt.Sprintf("%v", controlValue)
		if argsMap.Has(expectedIntKey) {
			expectedValue := argsMap.AsInt(ctx, expectedIntKey, 0)
			if controlValue != int32(expectedValue) {
				return errors.Reason("compare: expected %d is not equal to actual %d", expectedValue, controlValue).Err()
			} else {
				log.Debugf(ctx, "Compare (Int), actual and expected values are equals to %d", expectedValue)
			}
		} else if argsMap.Has(expectedIntGreaterKey) {
			expectedValue := argsMap.AsInt(ctx, expectedIntGreaterKey, 0)
			if controlValue > int32(expectedValue) {
				return errors.Reason("compare: expected value %d, actual value %d do not match", int32(expectedValue), controlValue).Err()
			} else {
				log.Debugf(ctx, "Compare (Int), expected value %s, actual value %d", expectedValue, controlValue)
			}
		} else if argsMap.Has(expectedIntLessKey) {
			expectedValue := argsMap.AsInt(ctx, expectedIntLessKey, 0)
			if controlValue < int32(expectedValue) {
				return errors.Reason("compare: expected value %d, actual value %d do not match", int32(expectedValue), controlValue).Err()
			} else {
				log.Debugf(ctx, "Compare (Int), expected value %s, actual value %d", expectedValue, controlValue)
			}
		}
	} else if argsMap.Has(expectedFloatKey) {
		expectedValue := argsMap.AsFloat64(ctx, expectedFloatKey, 0)
		controlValue, err := servodGetDouble(ctx, info.NewServod(), command)
		if err != nil {
			servodControlValue = err.Error()
			return errors.Annotate(err, "servo check servod control exec").Err()
		}
		servodControlValue = fmt.Sprintf("%v", controlValue)
		log.Debugf(ctx, "Compare (Double), expected value %s, actual value %f", expectedValue, controlValue)
		if controlValue != expectedValue {
			return errors.Reason("compare: expected value %f, actual value %f do not match", expectedValue, controlValue).Err()
		}
	} else if argsMap.Has(expectedBoolKey) {
		expectedValue := argsMap.AsBool(ctx, expectedBoolKey, false)
		controlValue, err := servodGetBool(ctx, info.NewServod(), command)
		if err != nil {
			servodControlValue = err.Error()
			return errors.Annotate(err, "servo check servod control exec").Err()
		}
		servodControlValue = fmt.Sprintf("%v", controlValue)
		log.Debugf(ctx, "Compare (Bool), expected value %s, actual value %t", expectedValue, controlValue)
		if controlValue != expectedValue {
			return errors.Reason("compare: expected value %t, actual value %t do not match", expectedValue, controlValue).Err()
		}
	} else {
		log.Infof(ctx, "Servo Check Servod Control Exec: expected value type not specified in config, or did not match any known types.")
		res, err := info.NewServod().Get(ctx, command)
		if err != nil {
			servodControlValue = err.Error()
			return errors.Annotate(err, "servo check servod control exec").Err()
		}
		// The value can contain different value types.
		// Ex.: "double:xxxx.xx"
		resRawString := strings.TrimSpace(res.String())
		servodControlValue = resRawString
		log.Infof(ctx, "Servo Check Servod Control Exec: for command %q, received %q.", command, resRawString)
	}
	return nil
}

func init() {
	execs.Register("servo_check_servod_control", servoCheckServodControlExec)
}
