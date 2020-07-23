// Copyright 2020 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// AUTOGENERATED. DO NOT EDIT.

// This file is generated by go.chromium.org/luci/tools/cmd/assets.
//
// It contains tests that ensure that assets embedded into the binary are
// identical to files on disk.

package templates

import (
	"go/build"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestAssets(t *testing.T) {
	t.Parallel()

	pkg, err := build.ImportDir(".", build.FindOnly)
	if err != nil {
		t.Fatalf("can't load package: %s", err)
	}

	fail := false
	for name := range Assets() {
		GetAsset(name) // for code coverage
		path := filepath.Join(pkg.Dir, filepath.FromSlash(name))
		blob, err := ioutil.ReadFile(path)
		if err != nil {
			t.Errorf("can't read file with assets %q (%s) - %s", name, path, err)
			fail = true
		} else if string(blob) != GetAssetString(name) {
			t.Errorf("embedded asset %q is out of date", name)
			fail = true
		}
	}

	if fail {
		t.Fatalf("run 'go generate' to update assets.gen.go")
	}
}
