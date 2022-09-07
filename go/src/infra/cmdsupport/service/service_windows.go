// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build windows
// +build windows

package service

import (
	"golang.org/x/sys/windows/svc"
)

// Implements svc.Handler
type serviceHandler struct {
	service     *Service
	returnValue int
}

func (m *serviceHandler) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	// https://cs.opensource.google/go/x/sys/+/20c2bfdb:windows/svc/service.go;l=57
	// "Note that Interrogate is always accepted"
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	changes <- svc.Status{State: svc.StartPending}

	// Start the service running in a new goroutine
	returnChan := make(chan int)
	go func() {
		// When service.Start() returns, will stop the loop below,
		// and set Windows Service state to "Stopped"
		// Should only return after receiving a "Stop" change request from Windows,
		// and calling the Service's provided Stop() function.
		returnChan <- m.service.Start()
	}()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

loop:
	for {
		select {
		// Listen for requests from Windows
		case changeRequest := <-r:
			switch changeRequest.Cmd {
			case svc.Interrogate:
				changes <- changeRequest.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				// User-provided handler to ask the service to stop
				m.service.Stop()
			default:
				// Unexpected control request
			}
		// Listen for service returning
		case returnValue := <-returnChan:
			// svc.Run() doesn't have a way to return an int - pass through serviceHandler instead
			m.returnValue = returnValue
			changes <- svc.Status{State: svc.Stopped}
			break loop
		}
	}
	return
}

func Run(s *Service) int {

	inService, err := svc.IsWindowsService()
	if err == nil && inService {
		// Running as a Windows Service
		handler := &serviceHandler{service: s}

		// svc.Run calls StartServiceCtrlDispatcher
		// https://docs.microsoft.com/en-us/windows/win32/api/winsvc/nf-winsvc-startservicectrldispatchera#remarks
		// * Must be called within 30 seconds of startup
		// * Name (first arg) is ignored if using "own process" mode (which we are)
		// * Will not return until the service enters the "Stopped" state.
		svc.Run("", handler)
		return handler.returnValue
	} else {
		// Running as a regular command
		return s.Start()
	}
}
