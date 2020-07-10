// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cli

import "go.chromium.org/luci/common/errors"

type outputFormat int

const (
	// protoJSON is JSON form of the chrome.dir_meta.Mapping protobuf message.
	protoJSON outputFormat = iota
	// chromeLegacy is the format used in
	// https://storage.googleapis.com/chromium-owners/component_map_subdirs.json
	chromeLegacy
)

func parseOutputFormat(format string) (outputFormat, error) {
	switch format {
	case "proto-json":
		return protoJSON, nil
	case "chrome-legacy":
		return chromeLegacy, nil
	default:
		return 0, errors.Reason(`unexpected format %q; valid values: "proto-json", "chrome-legacy"`, format).Err()
	}
}
