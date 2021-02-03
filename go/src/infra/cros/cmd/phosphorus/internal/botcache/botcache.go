// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package botcache provides an interface to interact with data cached in a
// swarming bot corresponding to a Chrome OS DUT.package botcache
package botcache

import (
	"infra/cros/cmd/phosphorus/internal/skylab_local_state/location"
	"path/filepath"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"go.chromium.org/chromiumos/infra/proto/go/lab_platform"
	"go.chromium.org/luci/common/errors"

	"os"
)

// Load reads the contents of the bot cache.
func Load(autotestDir string, fileID string) (*lab_platform.DutState, error) {
	p := cacheFilePath(autotestDir, fileID)
	s := lab_platform.DutState{}

	if err := readJSONPb(p, &s); err != nil {
		return nil, errors.Annotate(err, "load botcache").Err()
	}

	return &s, nil
}

// Save overwrites the contents of the bot cache with provided DutState.
func Save(autotestDir string, fileID string, s *lab_platform.DutState) error {
	p := location.CacheFilePath(autotestDir, fileID)

	if err := writeJSONPb(p, s); err != nil {
		return errors.Annotate(err, "write DUT state").Err()
	}

	return nil
}

func readJSONPb(inFile string, payload proto.Message) error {
	r, err := os.Open(inFile)
	if err != nil {
		return errors.Annotate(err, "read JSON pb").Err()
	}
	defer r.Close()

	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}
	if err := unmarshaler.Unmarshal(r, payload); err != nil {
		return errors.Annotate(err, "read JSON pb").Err()
	}
	return nil
}

func writeJSONPb(outFile string, payload proto.Message) error {
	dir := filepath.Dir(outFile)
	// Create the directory if it doesn't exist.
	if err := os.MkdirAll(dir, 0777); err != nil {
		return errors.Annotate(err, "write JSON pb").Err()
	}

	w, err := os.Create(outFile)
	if err != nil {
		return errors.Annotate(err, "write JSON pb").Err()
	}
	defer w.Close()

	marshaler := jsonpb.Marshaler{}
	if err := marshaler.Marshal(w, payload); err != nil {
		return errors.Annotate(err, "write JSON pb").Err()
	}
	return nil
}

const (
	botCacheSubDir     = "swarming_state"
	botCacheFileSuffix = ".json"
)

// cacheFilePath constructs the path to the state cache file.
func cacheFilePath(autotestDir string, fileID string) string {
	return filepath.Join(autotestDir, botCacheSubDir, fileID+botCacheFileSuffix)
}
