// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spec

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"infra/tools/vpython/api/env"

	"github.com/luci/luci-go/common/errors"

	"github.com/golang/protobuf/proto"
)

// Normalize normalizes the specification Message such that two messages
// with identical meaning will have the same identical representation.
func Normalize(spec *env.Spec) error {
	sort.Sort(envSpecPackageSlice(spec.Wheel))

	// No duplicate packages. Since we're sorted, we can just check for no
	// immediate repetitions.
	for i, pkg := range spec.Wheel {
		if i > 0 && pkg.Path == spec.Wheel[i-1].Path {
			return errors.Reason("duplicate spec entries for package %(path)q").
				D("path", pkg.Path).
				Err()
		}
	}

	return nil
}

// Hash hashes the contents of the supplied "spec" and returns the result as
// a hex-encoded string.
func Hash(spec *env.Spec) string {
	data, err := proto.Marshal(spec)
	if err != nil {
		panic(fmt.Errorf("failed to marshal proto: %v", err))
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

type envSpecPackageSlice []*env.Spec_Package

func (s envSpecPackageSlice) Len() int      { return len(s) }
func (s envSpecPackageSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s envSpecPackageSlice) Less(i, j int) bool {
	// Sort by Path, then Version.
	if s[i].Path >= s[j].Path {
		return false
	}
	return s[i].Version >= s[j].Version
}
