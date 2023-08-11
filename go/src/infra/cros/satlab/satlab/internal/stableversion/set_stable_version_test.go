// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stableversion

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
)

// TestWriteLocalStableVersion tests stable version file creation
func TestWriteLocalStableVersion(t *testing.T) {
	t.Parallel()

	rv := &RecoveryVersion{
		Board:   "zork",
		Model:   "gumboz",
		Os:      "R115-15474.70.0",
		Fw:      "Google_Berknip.13434.356.0",
		FwImage: "zork-firmware/R87-13434.819.0",
	}

	path := "./tmp/recovery_versions/"
	err := WriteLocalStableVersion(rv, path)
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

	rv2 := &RecoveryVersion{}
	_ = json.Unmarshal([]byte(savedRecoveryVersion), rv2)

	if *rv != *rv2 {
		t.Errorf("Recovery version saved incorrectly")
	}

}
