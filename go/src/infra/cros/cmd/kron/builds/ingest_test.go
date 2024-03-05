// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package builds

import "testing"

func TestExtractBoardAndVariant(t *testing.T) {
	amdTest := "amd64-generic"
	expectedBoard := amdTest
	expectedVariant := ""

	board, variant, err := extractBoardAndVariant(amdTest)
	if err != nil {
		t.Error(err)
		return
	}
	if board != expectedBoard {
		t.Errorf("expected %s got %s", expectedBoard, board)
	}
	if variant != expectedVariant {
		t.Errorf("expected %s got %s", amdTest, variant)
	}

	fizzLabstationTest := "fizz-labstation"
	expectedBoard = fizzLabstationTest
	expectedVariant = ""

	board, variant, err = extractBoardAndVariant(fizzLabstationTest)
	if err != nil {
		t.Error(err)
		return
	}
	if board != expectedBoard {
		t.Errorf("expected %s got %s", expectedBoard, board)
	}
	if variant != expectedVariant {
		t.Errorf("expected %s got %s", amdTest, variant)
	}

	test64 := "kevin64"
	expectedBoard = "kevin"
	expectedVariant = "64"

	board, variant, err = extractBoardAndVariant(test64)
	if err != nil {
		t.Error(err)
		return
	}
	if board != expectedBoard {
		t.Errorf("expected %s got %s", expectedBoard, board)
	}
	if variant != expectedVariant {
		t.Errorf("expected %s got %s", amdTest, variant)
	}

	test64Proper := "kevin-arc64"
	expectedBoard = "kevin"
	expectedVariant = "arc64"

	board, variant, err = extractBoardAndVariant(test64Proper)
	if err != nil {
		t.Error(err)
		return
	}
	if board != expectedBoard {
		t.Errorf("expected %s got %s", expectedBoard, board)
	}
	if variant != expectedVariant {
		t.Errorf("expected %s got %s", amdTest, variant)
	}

	testNormal := "test64-kernelnext"
	expectedBoard = "test64"
	expectedVariant = "kernelnext"

	board, variant, err = extractBoardAndVariant(testNormal)
	if err != nil {
		t.Error(err)
		return
	}
	if board != expectedBoard {
		t.Errorf("expected %s got %s", expectedBoard, board)
	}
	if variant != expectedVariant {
		t.Errorf("expected %s got %s", amdTest, variant)
	}
}
