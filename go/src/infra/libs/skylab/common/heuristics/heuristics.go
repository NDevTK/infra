// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package heuristics

import (
	"io/fs"
	"os"
	"regexp"
	"strings"

	"go.chromium.org/luci/common/errors"
)

// LooksLikeSatlabRemoteAccessContainer determines whether the container we are running on looks like
// a satlab remote access container.
func LooksLikeSatlabRemoteAccessContainer() (bool, error) {
	_, err := os.Stat("/usr/local/bin/get_host_identifier")
	if err == nil {
		// Happy path, we're in a satlab remote access container.
		return true, nil
	}
	if ok := errors.Is(err, fs.ErrNotExist); !ok {
		// Semi-happy path, we successfully determined that we're not in a satlab remote access container.
		return false, nil
	}
	return false, errors.Annotate(err, "looks like satlab remote access container").Err()
}

// LooksLikeSatlabDevice returns whether a hostname or botID appears to be a satlab-managed device.
// This function exists so that we use the same heuristic everywhere when identifying satlab devices.
func LooksLikeSatlabDevice(hostname string) bool {
	h := strings.TrimPrefix(hostname, "crossk-")
	return strings.HasPrefix(h, "satlab")
}

// LooksLikeLabstation returns whether a hostname or botID appears to be a labstation or not.
// This function exists so that we always use the same heuristic everywhere when identifying labstations.
func LooksLikeLabstation(hostname string) bool {
	return strings.Contains(hostname, "labstation")
}

// LooksLikeHeader heuristically determines whether a CSV line looks like
// a CSV header for the MCSV format.
func LooksLikeHeader(rec []string) bool {
	if len(rec) == 0 {
		return false
	}
	return strings.EqualFold(rec[0], "name")
}

var HwSwarmingBotIdPrefixes = []string{"crossk-", "cros-", "chrome-perf-"}

// NormalizeBotNameToDeviceName takes a bot name or a DUT name and normalizes it to a DUT name.
// The prefix "crossk-" or "cros-" is suspicious and should be removed. The suffix ".cros" is also suspicious and should be removed.
func NormalizeBotNameToDeviceName(hostname string) string {
	res := hostname
	for _, p := range HwSwarmingBotIdPrefixes {
		if strings.HasPrefix(hostname, p) {
			res = strings.TrimPrefix(hostname, p)
			break
		}
	}
	return strings.TrimSuffix(res, ".cros")
}

// looksLikeValidPool heuristically checks a string to see if it looks like a valid pool.
// A heuristically valid pool name contains only a-z, A-Z, 0-9, -, and _ .
// A pool name cannot begin with - and 0-9 .
var LooksLikeValidPool = regexp.MustCompile(`\A[A-Za-z_][-A-Za-z0-9_]*\z`).MatchString

// NormalizeTextualData lowercases data and removes leading and trailing whitespace.
func NormalizeTextualData(data string) string {
	return strings.ToLower(strings.TrimSpace(data))
}

// LooksLikeFieldMask checks whether a given string looks like a field mask.
var LooksLikeFieldMask = regexp.MustCompile(`\A[a-z][A-Za-z0-9\.]*\z`).MatchString

// TaskType is the high level type of a task such as repair or audit. It indicates what
// implementation (legacy or paris) and what level of stability (prod vs canary) should be used.
type TaskType int

const (
	// ProdTaskType refers to the paris task on the prod cipd label.
	ProdTaskType TaskType = 100
	// LatestTaskType refers to the paris task on the latest cipd label.
	LatestTaskType TaskType = 200
)
