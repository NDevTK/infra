// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"regexp"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
)

const (
	// crasAudioNodesQueryCmd query nodes information from Cras audio utilities using dbus-send.
	crasAudioNodesQueryCmd = "dbus-send --system --type=method_call --print-reply " +
		"--dest=org.chromium.cras /org/chromium/cras " +
		"org.chromium.cras.Control.GetNodes"
	// findCrasAudioNodeTypeRegexp is the regular expression that matches the string in the format of:
	// Ex:
	// string "Type"
	// variant             string "INTERNAL_MIC"
	findCrasAudioNodeTypeRegexp = `string "Type"\s+variant\s+string "%s"`
)

// crasAudioNodeTypeIsPlugged finds if the specified audio type node is present
// in the list of all plugged Cras Audio Nodes.
//
// Example of the type "INTERNAL_MIC" Cras Audio Nodes present on the DUT:
//
// dict entry(
//	string "Type"
// 	variant             string "INTERNAL_MIC"
// )
//
// @param nodeType : A string representing Cras Audio Node Type
// @returns: (true, nil) if the given nodeType is found in the output of the crasAudioNodesQueryCmd
//           (false, nil) if the given nodeType if not found in the output.
//           (false, err) if there is any error generated.
func crasAudioNodeTypeIsPlugged(ctx context.Context, r execs.Runner, nodeType string) (bool, error) {
	output, err := r(ctx, crasAudioNodesQueryCmd)
	if err != nil {
		return false, errors.Annotate(err, "node type of %s is plugged", nodeType).Err()
	}
	nodeTypeRegexp, err := regexp.Compile(fmt.Sprintf(findCrasAudioNodeTypeRegexp, nodeType))
	if err != nil {
		return false, errors.Annotate(err, "node type of %s is plugged", nodeType).Err()
	}
	nodeTypeExist := nodeTypeRegexp.MatchString(output)
	return nodeTypeExist, nil
}
