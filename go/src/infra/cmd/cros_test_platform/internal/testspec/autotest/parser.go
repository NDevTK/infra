// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package autotest

import (
	"fmt"
	"regexp"
	"strconv"

	"go.chromium.org/luci/common/errors"
)

func parseTestControl(text string) (*testMetadata, error) {
	var merr errors.MultiError
	var tm testMetadata
	merr = append(merr, parseName(text, &tm))
	merr = append(merr, parseSyncCount(text, &tm))
	merr = append(merr, parseRetries(text, &tm))
	return &tm, unwrapMultiErrorIfNil(merr)
}

func unwrapMultiErrorIfNil(merr errors.MultiError) error {
	if merr.First() == nil {
		return nil
	}
	return merr
}

func parseName(text string, tm *testMetadata) error {
	ms := namePattern.FindAllStringSubmatch(text, -1)
	var err error
	tm.Name, err = unwrapSingleSubmatchOrError(ms, "parseName")
	return err
}

func unwrapSingleSubmatchOrError(values [][]string, errContext string) (string, error) {
	switch len(values) {
	case 0:
		return "", nil
	case 1:
		m := values[0]
		if len(m) != 2 {
			// Number of sub-matches is determined only by the regexp
			// definitions in this module.
			// Incorrect number of sub-matches is thus a programming error that
			// will cause a panic() on *every* successful match.
			panic(fmt.Sprintf("%s: match has %d submatches, want 1: %s", errContext, len(m), m[1:]))
		}
		return m[1], nil
	default:
		return "", fmt.Errorf("%s: more than one value: %s", errContext, values)
	}
}

func parseSyncCount(text string, tm *testMetadata) error {
	ms := syncCountPattern.FindAllStringSubmatch(text, -1)
	sc, err := unwrapSingleSubmatchOrError(ms, "parseSyncCount")
	if err != nil {
		return err
	}
	if sc == "" {
		return nil
	}
	tm.DutCount, err = parseInt32OrError(sc, "parseSyncCount")
	if tm.DutCount > 0 {
		tm.NeedsMultipleDuts = true
	}
	return err
}

func parseInt32OrError(sc string, errContext string) (int32, error) {
	dc, err := strconv.ParseInt(sc, 10, 32)
	if err != nil {
		return 0, errors.Annotate(err, errContext).Err()
	}
	return int32(dc), nil
}

func parseRetries(text string, tm *testMetadata) error {
	ms := retriesPattern.FindAllStringSubmatch(text, -1)
	sc, err := unwrapSingleSubmatchOrError(ms, "parseRetries")
	if err != nil {
		return err
	}
	if sc == "" {
		setDefaultRetries(tm)
		return nil
	}
	tm.MaxRetries, err = parseInt32OrError(sc, "parseRetries")
	if tm.MaxRetries > 0 {
		tm.AllowRetries = true
	}
	return err
}

func setDefaultRetries(tm *testMetadata) {
	tm.AllowRetries = true
	tm.MaxRetries = 1
}

var (
	namePattern         = regexp.MustCompile(`\s*NAME\s*=\s*['"]([\w\.-]+)['"]\s*`)
	syncCountPattern    = regexp.MustCompile(`\s*SYNC_COUNT\s*=\s*(\d+)\s*`)
	retriesPattern      = regexp.MustCompile(`\s*JOB_RETRIES\s*=\s*(\d+)\s*`)
	dependenciesPattern = regexp.MustCompile(`\s*DEPENDENCIES\s*=\s*['"]([\w\.-]+)['"]\s*`)
)
