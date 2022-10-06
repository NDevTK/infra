// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/led/job"
	"google.golang.org/protobuf/types/known/structpb"
)

func (m tryRunBase) GetBuilderInputProps(ctx context.Context, fullBuilderName string) (*structpb.Struct, error) {
	bucket, builder, err := separateBucketFromBuilder(fullBuilderName)
	if err != nil {
		return &structpb.Struct{}, err
	}

	stdout, stderr, err := m.RunCmd(ctx, "led", "get-builder", fmt.Sprintf("%s:%s", bucket, builder))
	if err != nil {
		m.LogErr(stderr)
		return &structpb.Struct{}, err
	}

	var definition job.Definition_Buildbucket
	if err := json.Unmarshal([]byte(stdout), &definition); err != nil {
		// `led get-builder` returns proto enum fields as strings instead of ints, like buildbucket.bbagent_args.infra.experiment_reasons.
		// json.Unmarshal considers that to be an inappropriate type, returns an UnmarshalTypeError, and otherwise unmarshals as best as it can.
		// If the error is an UnmarshalTypeError, and InputProperties seem to have been unmarshaled OK, then there's no problem.
		inputProps := definitionToInputProperties(definition)
		if _, ok := err.(*json.UnmarshalTypeError); ok && inputProps != nil {
			return inputProps, nil
		}
		return &structpb.Struct{}, errors.Annotate(err, "unmarshaling `led get-builder` output").Err()
	}
	return definitionToInputProperties(definition), nil
}

// definitionToInputProperties extracts the input properties from an unmarshaled `led get-builder` output.
func definitionToInputProperties(definition job.Definition_Buildbucket) *structpb.Struct {
	return definition.Buildbucket.GetBbagentArgs().GetBuild().GetInput().GetProperties()
}

// writeStructToFile creates a tempfile and writes the struct as JSON data.
func writeStructToFile(s *structpb.Struct, file *os.File) error {
	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		return err
	}
	if _, err := file.Write(jsonBytes); err != nil {
		return err
	}
	return nil
}

// readStructFromFile reads a struct from the specified file.
func readStructFromFile(path string) (*structpb.Struct, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := &structpb.Struct{}
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return s, nil
}

func setProperty(s *structpb.Struct, key string, value interface{}) error {
	// Inner function for recursing over each component of the key ('.' separated).
	setPropertyInner := func(s *structpb.Struct, toProcess []string, value interface{}) error {
		processed := []string{}
		fields := s.Fields
		for k, v := range fields {
			if k == toProcess[0] {
				// Found the key.
				processed = append(processed, k)
				toProcess = toProcess[1:]
				// If there's no more key to process, assign the value.
				if len(toProcess) == 0 {
					structValue, err := structpb.NewValue(value)
					if err != nil {
						return errors.Annotate(err, "Could not convert value: ").Err()
					}
					fields[k] = structValue
					return nil
				}
				// If there's more key to process, recurse.
				if v.GetStructValue() == nil {
					return fmt.Errorf("The value for %s is not a struct, cannot resolve key %s.",
						strings.Join(processed, ","), k)
				} else {
					return setProperty(v.GetStructValue(), strings.Join(toProcess, "."), value)
				}
			}
		}
		// Couldn't find an existing key, create one.
		for len(toProcess) > 1 {
			var nextComponent string
			nextComponent, toProcess = toProcess[0], toProcess[1:]
			emptyStruct, err := structpb.NewStruct(map[string]interface{}{})
			if err != nil {
				return err
			}
			fields[nextComponent] = structpb.NewStructValue(emptyStruct)
			fields = fields[nextComponent].GetStructValue().Fields
		}
		structValue, err := structpb.NewValue(value)
		if err != nil {
			return errors.Annotate(err, "Could not convert value: ").Err()
		}
		fields[toProcess[0]] = structValue
		return nil
	}

	// If the value is a slice, make sure it's a slice of interfaces.
	// This way this function can be called with []string, []int, etc. values.
	if s := reflect.ValueOf(value); s.Kind() == reflect.Slice {
		if s.IsNil() {
			value = nil
		} else {
			slice := make([]interface{}, s.Len())
			for i := 0; i < s.Len(); i++ {
				slice[i] = s.Index(i).Interface()
			}
			value = slice
		}
	}

	return setPropertyInner(s, strings.Split(key, "."), value)
}
