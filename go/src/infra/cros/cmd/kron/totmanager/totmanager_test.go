// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package totmanager

import (
	"testing"

	suchpb "go.chromium.org/chromiumos/infra/proto/go/testplans"

	"infra/cros/internal/chromeosversion"
)

func TestInitTotManager(t *testing.T) {
	err := InitTotManager()
	if err != nil {
		t.Error(err)
	}

	if tot.ChromeBranch == 0 {
		t.Errorf("tot not properly fetched")
	}
}

func TestIsTargetedBranch(t *testing.T) {
	tot = chromeosversion.VersionInfo{
		ChromeBranch: 100,
	}

	totTest, err := IsTargetedBranch(100, []suchpb.Branch{suchpb.Branch_CANARY})
	if err != nil {
		t.Error(err)
		return
	}
	if !totTest {
		t.Errorf("expected %t got %t for TotTest", true, totTest)
		return
	}

	devTest, err := IsTargetedBranch(99, []suchpb.Branch{suchpb.Branch_DEV})
	if err != nil {
		t.Error(err)
		return
	}
	if !devTest {
		t.Errorf("expected %t got %t for devTest", true, devTest)
		return
	}

	betaTest, err := IsTargetedBranch(98, []suchpb.Branch{suchpb.Branch_BETA})
	if err != nil {
		t.Error(err)
		return
	}
	if !betaTest {
		t.Errorf("expected %t got %t for betaTest", true, betaTest)
		return
	}

	stableTest, err := IsTargetedBranch(97, []suchpb.Branch{suchpb.Branch_STABLE})
	if err != nil {
		t.Error(err)
		return
	}
	if !stableTest {
		t.Errorf("expected %t got %t for stableTest", true, stableTest)
		return
	}

	multiTest1, err := IsTargetedBranch(100, []suchpb.Branch{suchpb.Branch_CANARY, suchpb.Branch_STABLE})
	if err != nil {
		t.Error(err)
		return
	}
	if !multiTest1 {
		t.Errorf("expected %t got %t for multiTest1", true, multiTest1)
		return
	}

	multiTest2, err := IsTargetedBranch(97, []suchpb.Branch{suchpb.Branch_CANARY, suchpb.Branch_STABLE})
	if err != nil {
		t.Error(err)
		return
	}
	if !multiTest2 {
		t.Errorf("expected %t got %t for multiTest2", true, multiTest2)
		return
	}

	multiTest3, err := IsTargetedBranch(98, []suchpb.Branch{suchpb.Branch_CANARY, suchpb.Branch_STABLE})
	if err != nil {
		t.Error(err)
		return
	}
	if multiTest3 {
		t.Errorf("expected %t got %t for multiTest3", false, multiTest3)
		return
	}

}
