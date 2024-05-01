// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package resolver

import (
	"regexp"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
)

// PlaceholderRegex provides the format for how to find placeholders
// within the dynamic updates.
// Placeholders take the form of `${<placeholder>}`
// in which the named placeholder is wrapped by `${}`
//
// The only valid characters for the placeholder will be
// a combination of any letter, any number, and the `-` and `_`
// special characters.
const PlaceholderRegex = `\${[\w\d\-_]+}`

// Resolve converts the provided dynamic update into a json string
// and applies placeholder resolution, then converts back into a dynamic update object.
func Resolve(dynamicUpdate *api.UserDefinedDynamicUpdate, lookup map[string]string) (*api.UserDefinedDynamicUpdate, error) {
	jsonBytes, err := protojson.Marshal(dynamicUpdate)
	if err != nil {
		return nil, errors.Annotate(err, "failed to marshal dynamic update").Err()
	}

	resolvedJsonStr := ResolvePlaceholders(string(jsonBytes), lookup)
	// Run one more time. Allows placeholders to be embedded
	// within another placeholder by one, and only one, level.
	resolvedJsonStr = ResolvePlaceholders(resolvedJsonStr, lookup)

	resolvedUpdate := &api.UserDefinedDynamicUpdate{}
	unmarshaller := protojson.UnmarshalOptions{
		DiscardUnknown: true,
		AllowPartial:   true,
	}
	err = unmarshaller.Unmarshal([]byte(resolvedJsonStr), resolvedUpdate)
	if err != nil {
		return nil, errors.Annotate(err, "failed to unmarshal dynamic update").Err()
	}

	return resolvedUpdate, nil
}

// resolvePlaceholders searches the provided string for
// any placeholders and replaces them with the values
// corresponding in the lookup table.
//
// Placeholders should be a combination of letters, digits,
// hyphens, and underscores.
func ResolvePlaceholders(str string, lookup map[string]string) string {
	placeholders := regexp.MustCompile(PlaceholderRegex)
	return placeholders.ReplaceAllStringFunc(str, func(placeholder string) string {
		trimmed := strings.TrimLeft(placeholder, "${")
		lookupKey := strings.TrimRight(trimmed, "}")
		if value, ok := lookup[lookupKey]; ok {
			return value
		}
		return placeholder
	})
}
