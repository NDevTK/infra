// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package utils

import "errors"

var (
	ReachMaxRetry          = errors.New("reach the max retry")
	NotFound               = errors.New("can't find the value")
	EmptyQueue             = errors.New("the queue is empty")
	HostIdentifierNotFound = errors.New("can't get the host identifier")
)
