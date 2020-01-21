// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package querygs

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	labPlatform "go.chromium.org/chromiumos/infra/proto/go/lab_platform"
	"go.chromium.org/luci/common/gcloud/gs"

	"infra/libs/cros/stableversion/validateconfig"
)

const DONTCARE = "f7e8bdf6-f67c-4d63-aea3-46fa5e980403"

const exampleMetadataJSON = `
{
  "version": {
    "full": "R81-12835.0.0"
  },
  "results": [],
  "unibuild": true,
  "board-metadata": {
    "nami": {
      "models": {
        "sona": {
          "firmware-key-id": "SONA",
          "main-readonly-firmware-version": "Google_Nami.4.7.9",
          "ec-firmware-version": "nami_v1.2.3-4",
          "main-readwrite-firmware-version": "Google_Nami.42.43.44"
        },
        "akali360": {
          "firmware-key-id": "AKALI",
          "main-readonly-firmware-version": "Google_Nami.5.8.13",
          "ec-firmware-version": "nami_v1.2.3-4",
          "main-readwrite-firmware-version": "Google_Nami.52.53.54"
        }
      }
    }
  }
}
`

var testMaybeDownloadFileData = []struct {
	uuid     string
	metadata string
	out      map[string]map[string]string
}{
	{
		"f959c762-214e-4293-b655-032cd791a85f",
		exampleMetadataJSON,
		map[string]map[string]string{
			"nami": {
				"sona":     "Google_Nami.42.43.44",
				"akali360": "Google_Nami.52.53.54",
			},
		},
	},
}

func TestMaybeDownloadFile(t *testing.T) {
	t.Parallel()
	for _, tt := range testMaybeDownloadFileData {
		t.Run(tt.uuid, func(t *testing.T) {
			var r Reader
			r.dld = makeConstantDownloader(tt.metadata)
			e := r.maybeDownloadFile(DONTCARE, DONTCARE)
			if e != nil {
				msg := fmt.Sprintf("uuid (%s): unexpected error (%s)", tt.uuid, e.Error())
				t.Errorf(msg)
			}
			diff := cmp.Diff(tt.out, r.cache)
			if diff != "" {
				msg := fmt.Sprintf("uuid (%s): unexpected diff (%s)", tt.uuid, diff)
				t.Errorf(msg)
			}
		})
	}
}

var testFirmwareVersionData = []struct {
	name        string
	uuid        string
	metadata    string
	bt          string
	model       string
	CrOSVersion string
	out         string
}{
	{
		"extract firmware version \"sona\"",
		"20db725f-9f0e-457b-8da7-3fb6dfb57b7b",
		exampleMetadataJSON,
		"nami",
		"sona",
		"xxx-cros-version",
		"Google_Nami.42.43.44",
	},
	{
		"extract firmware version \"nami\"",
		"c9b5c1e3-a40f-4d5a-8353-6f890d80d9be",
		exampleMetadataJSON,
		"nami",
		"akali360",
		"xxx-cros-version",
		"Google_Nami.52.53.54",
	},
}

func TestGetFirmwareVersion(t *testing.T) {
	t.Parallel()
	for _, tt := range testFirmwareVersionData {
		t.Run(tt.name, func(t *testing.T) {
			var r Reader
			r.dld = makeConstantDownloader(tt.metadata)
			version, e := r.getFirmwareVersion(tt.bt, tt.model, tt.CrOSVersion)
			if e != nil {
				msg := fmt.Sprintf("name (%s): uuid (%s): unexpected error (%s)", tt.name, tt.uuid, e.Error())
				t.Errorf(msg)
			}
			diff := cmp.Diff(tt.out, version)
			if diff != "" {
				msg := fmt.Sprintf("name (%s): uuid (%s): unexpected diff (%s)", tt.name, tt.uuid, diff)
				t.Errorf(msg)
			}
		})
	}
}

var testValidateConfigData = []struct {
	name     string
	uuid     string
	metadata string
	in       string
	out      string
	hasErr   bool
}{
	{
		"empty",
		"b2b7aa51-3c2b-4c3f-93bf-0b26e6483489",
		`{}`,
		`{}`,
		`{
			"missing_boards": null,
			"failed_to_lookup": null,
			"invalid_versions": null
		}`,
		false,
	},
	{
		"two present boards",
		"f63476d1-0382-4098-8a15-18fcb1a2e61a",
		exampleMetadataJSON,
		`{
			"cros": [
				{
					"key": {
						"buildTarget": {"name": "nami"},
						"modelId": {}
					},
					"version": "R81-12835.0.0"
				}
			],
			"firmware": [
				{
					"key": {
						"modelId": {"value": "sona"},
						"buildTarget": {"name": "nami"}
					},
					"version": "Google_Nami.42.43.44"
				},
				{
					"key": {
						"modelId": {"value": "akali360"},
						"buildTarget": {"name": "nami"}
					},
					"version": "Google_Nami.52.53.54"
				}
			]
		}`,
		`{
			"missing_boards": null,
			"failed_to_lookup": null,
			"invalid_versions": null
		}`,
		false,
	},
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()
	for _, tt := range testValidateConfigData {
		t.Run(tt.name, func(t *testing.T) {
			var r Reader
			r.dld = makeConstantDownloader(tt.metadata)
			sv := parseStableVersionsOrPanic(tt.in)
			expected := parseResultsOrPanic(tt.out)
			result, e := r.ValidateConfig(sv)
			if tt.hasErr {
				if e == nil {
					msg := fmt.Sprintf("name (%s): uuid (%s): there should have been an error", tt.name, tt.uuid)
					t.Errorf(msg)
				}
			} else {
				if e != nil {
					msg := fmt.Sprintf("name (%s): uuid (%s): unexpected error (%s)", tt.name, tt.uuid, e.Error())
					t.Errorf(msg)
				}
			}
			diff := cmp.Diff(expected, result)
			if diff != "" {
				msg := fmt.Sprintf("name (%s): uuid (%s): unexpected diff (%s)", tt.name, tt.uuid, diff)
				t.Errorf(msg)
			}
		})
	}
}

var testRemoveWhiteListData = []struct {
	uuid string
	in   string
	out  string
}{
	{
		"bb30b384-38a6-4eb4-aa2c-b09815106d0a",
		"{}",
		"{}",
	},
	{
		"ef203c4d-6224-44df-ba80-9253dc47e4f7",
		`{
			"missing_boards": ["buddy_cfm"]
		}`,
		"{}",
	},
	{
		"41655030-dd58-481f-bcb4-4be4dfc01f07",
		`{
			"missing_boards": ["NOT A WHITELISTED BOARD"]
		}`,
		`{
			"missing_boards": ["NOT A WHITELISTED BOARD"]
		}`,
	},
	{
		"ff146760-fa2a-4495-aaea-42708843e6a1",
		`{
			"failed_to_lookup": [{"build_target": "fizz-labstation", "model": "fizz-labstation"}]
		}`,
		"{}",
	},
	{
		"24d51d50-8d34-498f-8c1d-b15341dfc549",
		`{
			"failed_to_lookup": [{"build_target": "NOT WHITELISTED", "model": "NOT WHITELISTED"}]
		}`,
		`{
			"failed_to_lookup": [{"build_target": "NOT WHITELISTED", "model": "NOT WHITELISTED"}]
		}`,
	},
}

func TestRemoveWhiteList(t *testing.T) {
	t.Parallel()
	for _, tt := range testRemoveWhiteListData {
		t.Run(tt.uuid, func(t *testing.T) {
			var in ValidationResult
			var out ValidationResult
			unmarshalOrPanic(tt.in, &in)
			unmarshalOrPanic(tt.out, &out)
			in.RemoveWhitelistedDUTs()
			if diff := cmp.Diff(out, in); diff != "" {
				msg := fmt.Sprintf("uuid (%s): unexpected diff (%s)", tt.uuid, diff)
				t.Errorf(msg)
			}
		})
	}
}

func makeConstantDownloader(content string) downloader {
	return func(gsPath gs.Path) ([]byte, error) {
		return []byte(content), nil
	}
}

// parseStableVersionsOrPanic is a helper function that's used in tests to feed
// a stable version file contained in a string literal to a test.
func parseStableVersionsOrPanic(content string) *labPlatform.StableVersions {
	out, err := validateconfig.ParseStableVersions([]byte(content))
	if err != nil {
		panic(err.Error())
	}
	return out
}

// parseStableVersionsOrPanic is a helper function that's used in tests to feed
// a result file contained in a string literal to a test.
func parseResultsOrPanic(content string) *ValidationResult {
	var out ValidationResult
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		panic(err.Error())
	}
	return &out
}

func unmarshalOrPanic(content string, dest interface{}) {
	if err := json.Unmarshal([]byte(content), dest); err != nil {
		panic(err.Error())
	}
}
