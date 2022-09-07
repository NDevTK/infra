// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package service allows running Go programs as Window Services.
//
// To run a program as a Windows Service, it must call `StartServiceCtrlDispatcher`
// and other Windows APIs to manage its state.
//
// The service package handles this for Windows, and is a no-op on other platforms.
//
// Based on the golang supplemental package example:
// https://pkg.go.dev/golang.org/x/sys/windows/svc/example
package service

// A Service is passed to Run, which will check if it is being run
// as a Service on Windows and call the Windows APIs as appropriate.
// If not, only the Start function will be called.
type Service struct {
	// Run the service. On Windows, will be in a new goroutine.
	// Return value will be returned by service.Run()
	Start func() int

	// Only used on Windows. Must cause the Start() function to return, running in another goroutine.
	Stop func()
}
