// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package server

import "testing"

func TestHasOwnership_match(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	hasOwnership := state.hasOwnership("a", "1")
	if !hasOwnership {
		t.Fatalf("hasOwnership should be true")
	}
}

func TestHasOwnership_nonExistOwnership(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	hasOwnership := state.hasOwnership("b", "2")
	if hasOwnership {
		t.Fatalf("hasOwnership should be false")
	}
}

func TestHasOwnership_updatedOwnership_matchCurrent(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	state.recordOwnership("b", "2")
	state.recordOwnership("a", "3")
	hasOwnership := state.hasOwnership("a", "3")
	if !hasOwnership {
		t.Fatalf("hasOwnership should reflect update")
	}
}

func TestHasOwnership_updatedOwnership_notMatchHistory(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	state.recordOwnership("b", "2")
	state.recordOwnership("a", "3")
	hasOwnership := state.hasOwnership("a", "1")
	if hasOwnership {
		t.Fatalf("hasOwnership should reflect update")
	}
}

func TestHasOwnership_cleared(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	state.recordOwnership("b", "2")
	state.recordOwnership("a", "3")
	state.clear()
	hasOwnership := state.hasOwnership("a", "3")
	if hasOwnership {
		t.Fatalf("hasOwnership should be cleared")
	}
}

func TestHasOwnership_removed_noMatch(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	state.recordOwnership("b", "2")
	state.removeOwnership("b")
	hasOwnership := state.hasOwnership("b", "2")
	if hasOwnership {
		t.Fatalf("hasOwnership should not match removed")
	}
}

func TestHasOwnership_removed_matchRemaining(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	state.recordOwnership("b", "2")
	state.removeOwnership("b")
	hasOwnership := state.hasOwnership("a", "1")
	if !hasOwnership {
		t.Fatalf("hasOwnership should match remaining")
	}
}

func TestOwnershipIdsReverseOrder_empty(t *testing.T) {
	state := newOwnershipState()
	ids := state.getIdsToClearOwnership()
	expect := []string{}
	check(t, ids, expect)
}

func TestOwnershipIdsReverseOrder_one(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	ids := state.getIdsToClearOwnership()
	expect := []string{"1"}
	check(t, ids, expect)
}

func TestOwnershipIdsReverseOrder_multiple(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	state.recordOwnership("b", "2")
	state.recordOwnership("c", "3")
	ids := state.getIdsToClearOwnership()
	expect := []string{"3", "2", "1"}
	check(t, ids, expect)
}

func TestOwnershipIdsReverseOrder_updatedOwnership(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	state.recordOwnership("b", "2")
	state.recordOwnership("b", "3")
	state.recordOwnership("a", "4")
	ids := state.getIdsToClearOwnership()
	expect := []string{"4", "3"}
	check(t, ids, expect)
}

func TestOwnershipIdsReverseOrder_cleared(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	state.recordOwnership("b", "2")
	state.recordOwnership("a", "3")
	state.clear()
	ids := state.getIdsToClearOwnership()
	expect := make([]string, 0)
	check(t, ids, expect)
}

func TestOwnershipIdsReverseOrder_removed(t *testing.T) {
	state := newOwnershipState()
	state.recordOwnership("a", "1")
	state.recordOwnership("b", "2")
	state.recordOwnership("c", "3")
	state.recordOwnership("b", "4")
	state.recordOwnership("a", "5")
	state.removeOwnership("b")
	ids := state.getIdsToClearOwnership()
	expect := []string{"5", "3"}
	check(t, ids, expect)
}

func TestRemove_nonExist(t *testing.T) {
	state := newOwnershipState()
	state.removeOwnership("b")
}
