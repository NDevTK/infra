// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package errors

import "errors"

var (
	NotMatch   = errors.New("can't match the value")
	RackExist  = errors.New("rack already added")
	AssetExist = errors.New("asset already added")
	DUTExist   = errors.New("DUT already added")
)
