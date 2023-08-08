// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package state

import (
	"strings"
	"testing"
)

func TestHasOwnership_match(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	hasOwnership := state.HasOwnership("a", "1")
	if !hasOwnership {
		t.Fatalf("HasOwnership should be true")
	}
}

func TestHasOwnership_nonExistOwnership(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	hasOwnership := state.HasOwnership("b", "2")
	if hasOwnership {
		t.Fatalf("HasOwnership should be false")
	}
}

func TestHasOwnership_updatedOwnership_matchCurrent(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	state.RecordOwnership("b", "2")
	state.RecordOwnership("a", "3")
	hasOwnership := state.HasOwnership("a", "3")
	if !hasOwnership {
		t.Fatalf("HasOwnership should reflect update")
	}
}

func TestHasOwnership_updatedOwnership_notMatchHistory(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	state.RecordOwnership("b", "2")
	state.RecordOwnership("a", "3")
	hasOwnership := state.HasOwnership("a", "1")
	if hasOwnership {
		t.Fatalf("HasOwnership should reflect update")
	}
}

func TestHasOwnership_cleared(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	state.RecordOwnership("b", "2")
	state.RecordOwnership("a", "3")
	state.Clear()
	hasOwnership := state.HasOwnership("a", "3")
	if hasOwnership {
		t.Fatalf("HasOwnership should be cleared")
	}
}

func TestHasOwnership_removed_noMatch(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	state.RecordOwnership("b", "2")
	state.RemoveOwnership("b")
	hasOwnership := state.HasOwnership("b", "2")
	if hasOwnership {
		t.Fatalf("HasOwnership should not match removed")
	}
}

func TestHasOwnership_removed_matchRemaining(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	state.RecordOwnership("b", "2")
	state.RemoveOwnership("b")
	hasOwnership := state.HasOwnership("a", "1")
	if !hasOwnership {
		t.Fatalf("HasOwnership should match remaining")
	}
}

func TestOwnershipIdsReverseOrder_empty(t *testing.T) {
	state := newOwnershipState()
	ids := state.GetIdsToClearOwnership()
	expect := []string{}
	check(t, ids, expect)
}

func TestOwnershipIdsReverseOrder_one(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	ids := state.GetIdsToClearOwnership()
	expect := []string{"1"}
	check(t, ids, expect)
}

func TestOwnershipIdsReverseOrder_multiple(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	state.RecordOwnership("b", "2")
	state.RecordOwnership("c", "3")
	ids := state.GetIdsToClearOwnership()
	expect := []string{"3", "2", "1"}
	check(t, ids, expect)
}

func TestOwnershipIdsReverseOrder_updatedOwnership(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	state.RecordOwnership("b", "2")
	state.RecordOwnership("b", "3")
	state.RecordOwnership("a", "4")
	ids := state.GetIdsToClearOwnership()
	expect := []string{"4", "3"}
	check(t, ids, expect)
}

func TestOwnershipIdsReverseOrder_cleared(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	state.RecordOwnership("b", "2")
	state.RecordOwnership("a", "3")
	state.Clear()
	ids := state.GetIdsToClearOwnership()
	expect := make([]string, 0)
	check(t, ids, expect)
}

func TestOwnershipIdsReverseOrder_removed(t *testing.T) {
	state := newOwnershipState()
	state.RecordOwnership("a", "1")
	state.RecordOwnership("b", "2")
	state.RecordOwnership("c", "3")
	state.RecordOwnership("b", "4")
	state.RecordOwnership("a", "5")
	state.RemoveOwnership("b")
	ids := state.GetIdsToClearOwnership()
	expect := []string{"5", "3"}
	check(t, ids, expect)
}

func TestRemove_nonExist(t *testing.T) {
	state := newOwnershipState()
	state.RemoveOwnership("b")
}

func check(t *testing.T, actual []string, expect []string) {
	actualStr := strings.Join(actual, ",")
	expectStr := strings.Join(expect, ",")
	if actualStr != expectStr {
		t.Fatalf("Slices do not match expect\nExpect: %v\nActual: %v", expect, actual)
	}
}
