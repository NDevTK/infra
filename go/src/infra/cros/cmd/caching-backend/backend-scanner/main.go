// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// backend-scanner is a service to scan caching backends and update a k8s
// configmap which mounted by caching frontend (i.e. the load balancer) and
// backends. It also scales the caching downloader replicas according to the
// count of backend server.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	port                 = flag.Int("port", 80, "the port the scanner service to listen")
	scanIntervalDuration = flag.Duration("scan-interval", 30*time.Minute, "the interval in seconds to scan all backends")

	namespace = flag.String("namespace", "skylab", "the k8s namespace of caching service")
	appLabel  = flag.String("app-label", "caching-backend", "the app label of the backend pods to scan")
)

func main() {
	if err := innerMain(); err != nil {
		log.Fatalf("Error: %s", err)
	}
	log.Printf("Done")
}

func innerMain() error {
	flag.Parse()

	s := &backendScanner{
		namespace: *namespace,
		appLabel:  *appLabel,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/update", s.updateHandler)
	svr := http.Server{Addr: fmt.Sprintf(":%d", *port), Handler: mux}
	log.Printf("starting backend scanner")

	if err := s.scanOnce(); err != nil {
		return err
	}
	log.Printf("Initial configmap created")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.regularlyScan(ctx, *scanIntervalDuration)
	log.Printf("Regular scanning scheduled at every %s", *scanIntervalDuration)

	err := svr.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	return err
}

type backendScanner struct {
	namespace string
	appLabel  string
}

func (s *backendScanner) updateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	if err := s.scanOnce(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.Printf("Error writing the response body: %s", err)
		}
	}
}

// scanOnce scans the specified caching backends and run the defined actions.
func (s *backendScanner) scanOnce() error {
	return nil
}

// regularlyScan scans at the specified interval duration.
func (s *backendScanner) regularlyScan(ctx context.Context, d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("Regular scanning stopped")
			return
		case <-ticker.C:
			log.Printf("Starting a regular scan")
			if err := s.scanOnce(); err != nil {
				log.Printf("The regular scan failed: %s", err)
			}
		}
	}
}
