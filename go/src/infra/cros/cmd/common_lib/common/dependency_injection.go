// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Inject implements the logic for which Dependency Injection is based on.
// Dependency injection is the process in which an incoming ProtoMessage can
// have a placeholder replaced with a value found within some provided
// map[string]interface{} which contains all the exposed values for dependency
// injection.
//
// Functionality:
// * Keys are split by "."
// * Keys should be camelCase
// * Keys can be a string or a number. Numbers would represent indexing an array.
// * If the injection_point is an array and the injectable is also an array of the same type, then it will override the array.
// * If the injection_point is an array and the injectable is not an array of the same type, then it will append into the array.
//
// Example:
//
//	message Example {
//		IpEndpoint endpoint = 1;
//	}
//
//	injectables := map[string]interface{}
//	injectables["example_endpoint"] = IpEndpoint {
//		address: "localhost",
//		port: 12345
//	}
//
//	receiver := Example {endpoint: {address: "localhost"}}
//
//	Inject(receiever, "endpoint.port", injectables, "example_endpoint.port")
func Inject(receiver protoreflect.ProtoMessage, injection_point string, injectables map[string]interface{}, injection_key string) (err error) {
	// Catch all thrown exceptions and recover to error instead.
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	receiver_map := ProtoToInterfaceMap(receiver)

	// Retrieve the injectable using the injection_key
	injectable := stepThroughInterface(injectables, strings.Split(injection_key, "."))

	// Cover 1-to-1 injections
	if injection_point == "" {
		receiver_map = injectable.(map[string]interface{})
	} else {
		// Isolate the injection_point and apply the injection
		injection_point_parts := strings.Split(injection_point, ".")
		point := stepThroughInterface(receiver_map, injection_point_parts[0:len(injection_point_parts)-1])

		// Special logic for last part of the injection_point as it may not exist
		// and therefore can't be indexed, as well as how to handle arrays.
		last_part := injection_point_parts[len(injection_point_parts)-1]
		if reflect.ValueOf(point.(map[string]interface{})[last_part]).Kind() == reflect.Slice {
			last_point := point.(map[string]interface{})[last_part]
			last_point = TranslateSliceToInterface(last_point)
			if reflect.ValueOf(last_point).Kind() == reflect.ValueOf(injectable).Kind() {
				point.(map[string]interface{})[last_part] = injectable
			} else {
				point.(map[string]interface{})[last_part] = append(last_point.([]interface{}), injectable)
			}
		} else {
			point.(map[string]interface{})[last_part] = injectable
		}
	}

	unmarshalInterfaceProtoMapToProto(receiver_map, receiver)

	return
}

// unmarshalInterfaceProtoMapToProto converts a map[string]interface{}
// back into a provided proto message.
func unmarshalInterfaceProtoMapToProto(proto_map map[string]interface{}, proto protoreflect.ProtoMessage) {
	json_bytes, _ := json.Marshal(proto_map)
	err := protojson.Unmarshal(json_bytes, proto)
	if err != nil {
		panic(err)
	}
}

// protoToInterfaceMap converts a proto message into an interface map
// which is a useful struct for dependency injection.
func ProtoToInterfaceMap(proto protoreflect.ProtoMessage) map[string]interface{} {
	json_bytes, err := protojson.Marshal(proto)
	if err != nil {
		panic(err)
	}
	var proto_map map[string]interface{}
	err = json.Unmarshal(json_bytes, &proto_map)
	if err != nil {
		panic(err)
	}

	return proto_map
}

// stepThroughInterface uses an array of keys to step through
// the indexes of the provided interface{}.
func stepThroughInterface(object interface{}, steps []string) interface{} {
	curr := object
	for _, step := range steps {
		if step_num, err := strconv.ParseInt(step, 10, 64); err == nil {
			curr = TranslateSliceToInterface(curr)
			curr = curr.([]interface{})[step_num]
		} else {
			curr = curr.(map[string]interface{})[step]
		}
	}

	return curr
}

// TranslateSliceToInterface expects a slice object in.
// The slice then gets forcefully casted into []interface{}
// which is a generic slice form that can be interacted with
// by dependency injection.
func TranslateSliceToInterface(slice interface{}) []interface{} {
	_type := reflect.ValueOf(slice)
	if _type.IsNil() {
		return nil
	}
	if _type.Kind() != reflect.Slice {
		panic(errors.New(fmt.Sprintf("Cannot translate %s objects to []interface{}", _type.Kind())))
	}

	result := make([]interface{}, _type.Len())
	for i := range result {
		result[i] = _type.Index(i).Interface()
	}

	return result
}
