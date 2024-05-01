// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
)

// CheckIfFieldDefinitionExists checks if a field definition exist in the provided msg type
func CheckIfFieldDefinitionExists(msg proto.Message, fieldName string) (bool, error) {
	dynMsg := dynamicpb.NewMessage(msg.ProtoReflect().Descriptor())
	msg1, err := proto.Marshal(msg)
	if err != nil {
		return false, errors.Annotate(err, "failed while marshaling provided proto").Err()
	}
	err = proto.Unmarshal(msg1, dynMsg)
	if err != nil {
		return false, errors.Annotate(err, "failed while unmarshalling dynamic proto").Err()
	}

	return dynMsg.Descriptor().Fields().ByName(protoreflect.Name(fieldName)) != nil, nil
}

// CheckIfSchedulingUnitFieldExistsInSuiteMD checks if scheduling_units field
// is defined in suiteMD.
// TODO(azrahman): remove this in future when it's safe.
func CheckIfSchedulingUnitFieldExistsInSuiteMD() (bool, error) {
	newFieldName := "scheduling_units"
	proto := &api.SuiteMetadata{}
	return CheckIfFieldDefinitionExists(proto, newFieldName)
}
