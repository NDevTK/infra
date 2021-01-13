// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// +build !windows

package fleetsignal

import (
	"log"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

// NotifySIGTERM monitors if any OS signals are received.
func NotifySIGTERM() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, unix.SIGINT, unix.SIGHUP, unix.SIGTERM, unix.SIGQUIT)
	sig := <-sigChan
	log.Printf("Captured %v, stopping fleet-tlw service and cleaning up...", sig)
}
