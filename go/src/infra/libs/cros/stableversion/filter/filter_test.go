// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package filter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"infra/libs/cros/stableversion/validateconfig"

	labPlatform "go.chromium.org/chromiumos/infra/proto/go/lab_platform"
)

var testByModelData = []struct {
	name  string
	in    *labPlatform.StableVersions
	model string
	out   *labPlatform.StableVersions
}{
	{
		"empty container",
		&labPlatform.StableVersions{},
		"",
		&labPlatform.StableVersions{},
	},
	{
		"empty container with nontrivial model",
		&labPlatform.StableVersions{},
		"vayne",
		&labPlatform.StableVersions{},
	},
	{
		"leave unaffected model",
		parseStableVersionsOrPanic(`
{
        "cros": [
                {
                        "key": {
                                "modelId": {

                                },
                                "buildTarget": {
                                        "name": "nami"
                                }
                        },
                        "version": "R81-12871.41.0"
                }
        ],
        "firmware": [
                {
                        "key": {
                                "modelId": {
                                        "value": "vayne"
                                },
                                "buildTarget": {
                                        "name": "nami"
                                }
                        },
                        "version": "Google_Nami.10775.123.0"
                }
        ]
}`),
		"xxxxxx",
		&labPlatform.StableVersions{},
	},
	{
		"retain explicitly named model",
		parseStableVersionsOrPanic(`
{
        "cros": [
                {
                        "key": {
                                "modelId": {

                                },
                                "buildTarget": {
                                        "name": "nami"
                                }
                        },
                        "version": "R81-12871.41.0"
                }
        ],
        "firmware": [
                {
                        "key": {
                                "modelId": {
                                        "value": "vayne"
                                },
                                "buildTarget": {
                                        "name": "nami"
                                }
                        },
                        "version": "Google_Nami.10775.123.0"
                }
        ]
}`),
		"vayne", parseStableVersionsOrPanic(`{
        "cros": [
                {
                        "key": {
                                "modelId": {

                                },
                                "buildTarget": {
                                        "name": "nami"
                                }
                        },
                        "version": "R81-12871.41.0"
                }
        ],
        "firmware": [
                {
                        "key": {
                                "modelId": {
                                        "value": "vayne"
                                },
                                "buildTarget": {
                                        "name": "nami"
                                }
                        },
                        "version": "Google_Nami.10775.123.0"
                }
        ]
}`),
	},
	{
		"keep everything with empty model",
		parseStableVersionsOrPanic(`
{
        "cros": [
                {
                        "key": {
                                "modelId": {

                                },
                                "buildTarget": {
                                        "name": "nami"
                                }
                        },
                        "version": "R81-12871.41.0"
                }
        ],
        "firmware": [
                {
                        "key": {
                                "modelId": {
                                        "value": "vayne"
                                },
                                "buildTarget": {
                                        "name": "nami"
                                }
                        },
                        "version": "Google_Nami.10775.123.0"
                }
        ]
}`),
		"",
		parseStableVersionsOrPanic(`{
        "cros": [
                {
                        "key": {
                                "modelId": {

                                },
                                "buildTarget": {
                                        "name": "nami"
                                }
                        },
                        "version": "R81-12871.41.0"
                }
        ],
        "firmware": [
                {
                        "key": {
                                "modelId": {
                                        "value": "vayne"
                                },
                                "buildTarget": {
                                        "name": "nami"
                                }
                        },
                        "version": "Google_Nami.10775.123.0"
                }
        ]
}`),
	},
}

// errorStartsWithDWIM checks if the error's message starts with a given prefix.
// If the prefix is empty, it checks whether the error is nil.
func errorStartsWithDWIM(err error, prefix string) bool {
	if prefix == "" {
		return err == nil
	}
	return strings.HasPrefix(err.Error(), prefix)
}

// errorWithDefault is a helper function for testing against error strings.
func errorWithDefault(e error, def string) string {
	if e == nil {
		return def
	}
	return e.Error()
}

func TestByModel(t *testing.T) {
	t.Parallel()
	for _, tt := range testByModelData {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := ByModel(tt.in, tt.model)
			if err != nil {
				t.Errorf("unexpected failure: %s", err)
			}
			if diff := cmp.Diff(tt.out, out); diff != "" {
				msg := fmt.Sprintf("name: %s, diff: %s", tt.name, diff)
				t.Errorf("%s", msg)
			}
		})
	}
}

// parseStableVersionsOrPanic is a helper function that's used in tests to feed
// a stable version file contained in a string literal to a test.
func parseStableVersionsOrPanic(contents string) *labPlatform.StableVersions {
	out, err := validateconfig.ParseStableVersions([]byte(contents))
	if err != nil {
		panic(err.Error())
	}
	return out
}
