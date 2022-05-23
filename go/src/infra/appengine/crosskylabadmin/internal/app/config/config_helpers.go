// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"context"
	"regexp"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// PermilleData contains information on what portion of traffic to opt
// into Prod and Latest.
type PermilleData struct {
	Source string
	Prod   float64
	Latest float64
}

func validatePattern(pattern string) error {
	if len(pattern) < 2 {
		return errors.Reason("pattern is too short").Err()
	}
	if pattern == "^$" {
		return errors.Reason(`pattern "^$" is not useful`).Err()
	}
	if pattern[0] == '^' || pattern[len(pattern)-1] == '$' {
		return nil
	}
	return errors.Reason(`pattern missing "^" or "$" anchor`).Err()
}

// matches returns true if and only if pattern and hostname are both nonempty and
// the hostname is an instance of the pattern.
//
// We return an non-nil error if and only if the pattern is empty, the hostname is empty,
// or the regular expression pattern fails to compile.
func matches(pattern string, hostname string) (bool, error) {
	if err := validatePattern(pattern); err != nil {
		return false, errors.Annotate(err, "matches").Err()
	}
	if hostname == "" {
		return false, errors.Reason("matches: hostname cannot be empty").Err()
	}
	r, err := regexp.Compile(pattern)
	if err != nil {
		return false, errors.Annotate(err, "matches").Err()
	}
	return r.MatchString(hostname), nil
}

// getLastMatch returns the last match in the config.
func (x *RolloutConfig) getLastMatch(hostname string) (*PermilleData, error) {
	patterns := x.GetPattern()
	// We want to enable the users to write the patterns with more general patterns at the top
	// and more specific patterns at the bottom.
	//
	// In this setting, it is correct to iterate backwards through the list and stop as soon
	// as we see a match.
	for i := -1 + len(patterns); i >= 0; i-- {
		ok, err := matches(patterns[i].GetPattern(), hostname)
		if err != nil {
			return nil, errors.Annotate(err, "get specific pattern").Err()
		}

		if ok {
			return &PermilleData{
				Source: patterns[i].GetPattern(),
				Prod:   float64(patterns[i].GetProdPermille()),
				Latest: float64(patterns[i].GetLatestPermille()),
			}, nil
		}
	}
	return &PermilleData{
		Source: "",
		Prod:   float64(x.GetProdPermille()),
		Latest: float64(x.GetLatestPermille()),
	}, nil
}

// ComputeProdPermille computes the most applicable prod permille
// for a device.
func (x *RolloutConfig) ComputePermilleData(ctx context.Context, hostname string) (*PermilleData, error) {
	d, err := x.getLastMatch(hostname)
	if err != nil {
		logging.WithError(err).Errorf(ctx, "malformed config file")
	}
	return d, nil
}
