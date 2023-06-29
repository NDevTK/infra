// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package testutils

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ShouldEqualProto compares two protobuf messages and return the result.
func ShouldEqualProto(actual interface{}, expected ...interface{}) string {
	if message := need(1, expected); message != success {
		return message
	}

	x, ok := actual.(protoreflect.ProtoMessage)
	if !ok {
		return fmt.Sprintf("Actual is not a proto message: %s", actual)
	}
	y, ok := expected[0].(protoreflect.ProtoMessage)
	if !ok {
		return fmt.Sprintf("Expected is not a proto message: %s", expected[0])
	}

	if ok := proto.Equal(x, y); !ok {
		return fmt.Sprintf("Actual: %s\nExpected: %s", x, y)
	}

	return success
}

// From goconvey
const (
	success                = ""
	needExactValues        = "This assertion requires exactly %d comparison values (you provided %d)."
	needNonEmptyCollection = "This assertion requires at least 1 comparison value (you provided 0)."
	needFewerValues        = "This assertion allows %d or fewer comparison values (you provided %d)."
)

func need(needed int, expected []interface{}) string {
	if len(expected) != needed {
		return fmt.Sprintf(needExactValues, needed, len(expected))
	}
	return success
}
