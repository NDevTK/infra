// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stableversion

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"infra/cros/recovery/models"
)

// TestWriteLocalStableVersion tests stable version file creation
func TestWriteLocalStableVersion(t *testing.T) {
	t.Parallel()

	rv := &models.RecoveryVersion{
		Board:     "zork",
		Model:     "gumboz",
		OsImage:   "R115-15474.70.0",
		FwVersion: "Google_Berknip.13434.356.0",
		FwImage:   "zork-firmware/R87-13434.819.0",
	}

	// Perform our test in a temporary file managed by the test framework.
	path := filepath.Join(t.TempDir(), "tmp", "recovery_versions")

	os.RemoveAll(path)
	err := os.MkdirAll(path, 0777)
	if err != nil {
		t.Errorf("Unexpected err: %v", err)
	}

	err = writeLocalStableVersion(rv, path)
	if err != nil {
		t.Errorf("Unexpected err: %v", err)
	}
	file := fmt.Sprintf("%s%s-%s.json", path, rv.Board, rv.Model)
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		t.Errorf("Unexpected err: %v", err)
	}

	savedRecoveryVersion, err := os.ReadFile(file)
	if err != nil {
		t.Errorf("Unexpected err: %v", err)
	}

	rv2 := &models.RecoveryVersion{}
	_ = json.Unmarshal([]byte(savedRecoveryVersion), rv2)

	if !reflect.DeepEqual(rv, rv2) {
		t.Errorf("Recovery version saved incorrectly")
	}
}

func TestValidateBoardModelArgs(t *testing.T) {
	var tests = []struct {
		name string
		sv   setStableVersionRun
		resp int
		err  bool
	}{
		{"All required params", setStableVersionRun{board: "zork", model: "gumboz", os: "R115-15474.70.0", fw: "Google_Berknip.13434.356.0", fwImage: "zork-firmware/R87-13434.819.0"}, 3, false},
		{"Minimal required params", setStableVersionRun{board: "zork", model: "gumboz"}, 0, false},
		{"Lack of a board name", setStableVersionRun{model: "gumboz", os: "R115-15474.70.0", fw: "Google_Berknip.13434.356.0", fwImage: "zork-firmware/R87-13434.819.0"}, 0, true},
		{"Lack of a model name", setStableVersionRun{board: "zork", os: "R115-15474.70.0", fw: "Google_Berknip.13434.356.0", fwImage: "zork-firmware/R87-13434.819.0"}, 0, true},
		{"Using flex as a Partner", setStableVersionRun{board: "zork", model: "gumboz", os: "R115-15474.70.0", fw: "Google_Berknip.13434.356.0", fwImage: "zork-firmware/R87-13434.819.0", isFlex: true}, 0, true},
		{"Partial version info", setStableVersionRun{board: "zork", model: "gumboz", os: "R115-15474.70.0", fwImage: "zork-firmware/R87-13434.819.0"}, 2, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ans, err := tt.sv.validateBoardModelArgs()
			if tt.err && err == nil {
				t.Fatal("want an error for invalid args")
			}
			if !tt.err && err != nil {
				t.Fatalf("got an error (%s) for valid args", err.Error())
			}
			if ans != tt.resp {
				t.Errorf("got (%d), want (%d)", ans, tt.resp)
			}
		})
	}
}

func TestValidateHostnameArgs(t *testing.T) {
	var tests = []struct {
		name string
		sv   setStableVersionRun
		err  bool
	}{
		{"All required params", setStableVersionRun{hostname: "satlab-11111111-host1", os: "R115-15474.70.0", fw: "Google_Berknip.13434.356.0", fwImage: "zork-firmware/R87-13434.819.0"}, false},
		{"Lack of hostname", setStableVersionRun{os: "R115-15474.70.0", fw: "Google_Berknip.13434.356.0", fwImage: "zork-firmware/R87-13434.819.0"}, true},
		{"Only hostname", setStableVersionRun{hostname: "satlab-11111111-host1"}, true},
		{"Only hostname with os", setStableVersionRun{hostname: "satlab-11111111-host1", os: "R115-15474.70.0"}, true},
		{"Hostname with os and flex", setStableVersionRun{hostname: "satlab-11111111-host1", os: "R115-15474.70.0", isFlex: true}, false},
		{"Flex with fw version", setStableVersionRun{hostname: "satlab-11111111-host1", os: "R115-15474.70.0", fw: "Google_Berknip.13434.356.0", isFlex: true}, true},
		{"Flex with fwImage version", setStableVersionRun{hostname: "satlab-11111111-host1", os: "R115-15474.70.0", fwImage: "zork-firmware/R87-13434.819.0", isFlex: true}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.sv.validateHostnameArgs()
			if tt.err && err == nil {
				t.Fatal("want an error for invalid args")
			}
			if !tt.err && err != nil {
				t.Errorf("got an error (%s) for valid args", err.Error())
			}
		})
	}
}
