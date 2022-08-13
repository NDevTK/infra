// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package swarming

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"go.chromium.org/chromiumos/config/go/payload"
	"go.chromium.org/chromiumos/config/go/test/api"
)

var LabelMarshaler = jsonpb.Marshaler{
	EnumsAsInts:  false,
	EmitDefaults: true,
	Indent:       "  ",
	OrigName:     true,
}

// ConvertAll converts one DutAttribute label to multiple Swarming labels.
//
// The converted labels are returned in the form of `${label_name}:val1,val2`
// in an array. Each label value is comma-separated. Label labelNames are the
// DutAttribute ID and the aliases listed.
func ConvertAll(dutAttr *api.DutAttribute, flatConfig *payload.FlatConfig) ([]string, error) {
	labelNames, err := GetLabelNames(dutAttr)
	if err != nil {
		return nil, err
	}

	// Construct and try each path defined in DutAttribute. Tried in order. First
	// path to return a value will be used.
	jsonPaths, err := ConstructJsonPaths(dutAttr)
	if err != nil {
		return nil, err
	}

	for _, p := range jsonPaths {
		valuesStr, err := GetLabelValuesStr(p, flatConfig)
		if err != nil {
			return nil, err
		}
		if err == nil && valuesStr != "" {
			return FormLabels(labelNames, valuesStr)
		}
	}
	return nil, errors.New("no supported config source found")
}

// FormLabels pairs label names with the label values `${label_name}:val1,val2`.
func FormLabels(labelNames []string, valuesStr string) ([]string, error) {
	// Exhausted all possible paths defined in DutAttribute. If valuesStr is empty,
	// then no values found.
	if valuesStr == "" {
		return nil, errors.New("no label values found in config source found")
	}

	var labels []string
	for _, n := range labelNames {
		labels = append(labels, fmt.Sprintf("%s:%s", n, valuesStr))
	}
	if len(labels) == 0 {
		return nil, errors.New("no labels can be generated")
	}
	return labels, nil
}

// GetLabelNames extracts all possible label names from a DutAttribute.
//
// For each DutAttribute, the main label name is defined by its ID value. In
// addition, users can define other aliases. GetLabelNames will return all as
// valid label names. The first label is always the main label as defined by the
// ID value.
func GetLabelNames(dutAttr *api.DutAttribute) ([]string, error) {
	name := dutAttr.GetId().GetValue()
	if name == "" {
		return nil, errors.New("DutAttribute has no ID")
	}
	return append([]string{name}, dutAttr.GetAliases()...), nil
}

// GetLabelValuesStr takes a path and returns the proto value.
//
// It uses a jsonpath string to try to find corresponding values in a proto. It
// returns a comma-separated string of the values found.
func GetLabelValuesStr(jsonGetPath string, pm proto.Message) (string, error) {
	js, err := LabelMarshaler.MarshalToString(pm)
	if err != nil {
		return "", err
	}

	pmJson := interface{}(nil)
	err = json.Unmarshal([]byte(js), &pmJson)
	if err != nil {
		return "", err
	}

	labelVals, err := jsonpath.Get(jsonGetPath, pmJson)
	if err != nil {
		return "", err
	}
	return ConstructLabelValuesString(labelVals)
}

// ConstructLabelValuesString takes label values and returns them as a string.
//
// It takes an interface of label values parsed from a json object and returns a
// comma-separated string of the values. The interfaces supported are primitive
// types and iterable interfaces.
func ConstructLabelValuesString(labelVals interface{}) (string, error) {
	var rsp string
	switch x := labelVals.(type) {
	case []interface{}:
		valsArr := []string{}
		for _, i := range x {
			i, ok := i.(string)
			if !ok {
				return "", fmt.Errorf("cannot cast to string: %s", i)
			}
			valsArr = append(valsArr, i)
		}
		rsp = strings.Join(valsArr, ",")
	case bool:
		rsp = strconv.FormatBool(labelVals.(bool))
	case float64:
		rsp = strconv.FormatFloat(labelVals.(float64), 'f', -1, 64)
	default:
		var ok bool
		rsp, ok = labelVals.(string)
		if !ok {
			return "", fmt.Errorf("cannot cast to string: %s", rsp)
		}
	}
	return rsp, nil
}

// ConstructJsonPaths returns config paths defined by a DutAttribute.
//
// It takes a DutAttribute and returns an array of field paths defined in
// jsonpath syntax. The sources that are currently supported are:
//  1. FlatConfigSource
//  2. HwidSource
func ConstructJsonPaths(dutAttr *api.DutAttribute) ([]string, error) {
	if dutAttr.GetFlatConfigSource() != nil {
		return generateFlatConfigSourcePaths(dutAttr), nil
	} else if dutAttr.GetHwidSource() != nil {
		return generateHwidSourcePaths(dutAttr), nil
	}
	return []string{}, errors.New("no supported config source found")
}

// generateFlatConfigSourcePaths returns config paths defined by a DutAttribute.
//
// It takes a DutAttribute and returns an array of FlatConfigSource field paths
// strings defined in jsonpath syntax.
func generateFlatConfigSourcePaths(dutAttr *api.DutAttribute) []string {
	var rsp []string
	for _, f := range dutAttr.GetFlatConfigSource().GetFields() {
		rsp = append(rsp, fmt.Sprintf("$.%s", f.GetPath()))
	}
	return rsp
}

// generateHwidSourcePaths returns config paths defined by a DutAttribute.
//
// It takes a DutAttribute and returns an array of HwidSource field paths
// strings defined in jsonpath syntax.
func generateHwidSourcePaths(dutAttr *api.DutAttribute) []string {
	var rsp []string
	componentType := dutAttr.GetHwidSource().GetComponentType()
	for _, f := range dutAttr.GetHwidSource().GetFields() {
		rsp = append(rsp, fmt.Sprintf(`$.hw_components[?(@.%s != null)].%s`, componentType, f.GetPath()))
	}
	return rsp
}
