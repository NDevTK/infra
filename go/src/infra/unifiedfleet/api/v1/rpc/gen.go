// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package ufspb contains the fleet service API.
package ufspb

//go:generate cproto -proto-path ../../../../..
//go:generate svcdec -type FleetServer

// Define a mock client here for the services that need a mock UFS.
//go:generate mockgen -copyright_file copyright.txt -source fleet.pb.go -destination mock/client.mock.go -package mockufs
