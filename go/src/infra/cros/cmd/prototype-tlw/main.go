// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Command dev-tlw implements the TLS wiring API for development convenience.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

var (
	port = flag.Int("port", 0, "Port to listen to")
	lab8 = flag.Bool("lab8", false, "Use caching server in chromeos8 lab")
)

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("dev-tlw: %s", err)
	}
	s := server{}
	s.lab8 = *lab8
	if err := s.Serve(l); err != nil {
		log.Fatalf("dev-tlw: %s", err)
	}
}
