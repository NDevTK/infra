// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Siso is a Ninja-compatible build system optimized for remote execution.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof" // import to let pprof register its HTTP handlers
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	log "github.com/golang/glog"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/client/versioncli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/system/signals"

	"infra/build/siso/auth/cred"
	"infra/build/siso/subcmd/ninja"
)

var (
	pprofAddr  string
	cpuprofile string
	memprofile string
	traceFile  string
)

const version = "0.1"

func getApplication() *cli.Application {
	authOpts := cred.AuthOpts()

	return &cli.Application{
		Name:  "siso",
		Title: "Ninja-compatible build system optimized for remote execution",
		Context: func(ctx context.Context) context.Context {
			ctx, cancel := context.WithCancel(ctx)
			signals.HandleInterrupt(cancel)()
			return ctx
		},
		Commands: []*subcommands.Command{
			subcommands.CmdHelp,

			ninja.Cmd(authOpts),

			authcli.SubcommandInfo(authOpts, "whoami", false),
			authcli.SubcommandLogin(authOpts, "login", false),
			authcli.SubcommandLogout(authOpts, "logout", false),
			versioncli.CmdVersion(version),
		},
	}
}

func main() {
	// Wraps sisoMain() because os.Exit() doesn't wait defers.
	os.Exit(sisoMain())
}

func sisoMain() int {
	// TODO(b/274361523): Ensure that these flags show up in `siso help`.
	flag.StringVar(&pprofAddr, "pprof_addr", "", `listen address for "go tool pprof". e.g. "localhost:6060"`)
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to this file")
	flag.StringVar(&memprofile, "memprofile", "", "write memory profile to this file")
	flag.StringVar(&traceFile, "trace", "", "go trace output for `go tool trace`")
	flag.Parse()

	// Flush the log on exit to not lose any messages.
	defer log.Flush()

	// Print a stack trace when a panic occurs.
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Fatalf("panic: %v\n%s", r, buf)
		}
	}()

	// Start an HTTP server that can be used to profile Siso during runtime.
	if pprofAddr != "" {
		// https://pkg.go.dev/net/http/pprof
		fmt.Fprintf(os.Stderr, "pprof is enabled, listening at http://%s/debug/pprof/\n", pprofAddr)
		go func() {
			log.Infof("pprof http listener: %v", http.ListenAndServe(pprofAddr, nil))
		}()
		defer func() {
			fmt.Fprintf(os.Stderr, "pprof is still listening at http://%s/debug/pprof/\n", pprofAddr)
			fmt.Fprintln(os.Stderr, "Press Ctrl-C to terminate the process")
			sigch := make(chan os.Signal, 1)
			signal.Notify(sigch, signals.Interrupts()...)
			<-sigch
		}()
	}

	// Save a CPU profile to disk on exit.
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatalf("failed to create cpuprofile file: %v", err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Errorf("failed to start CPU profiler: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Save a heap profile to disk on exit.
	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatalf("failed to create memprofile file: %v", err)
		}
		defer func() {
			err := pprof.WriteHeapProfile(f)
			if err != nil {
				log.Errorf("failed to write heap profile: %v", err)
			}
		}()
	}

	// Save a go trace to disk during execution.
	if traceFile != "" {
		fmt.Fprintf(os.Stderr, "enable go trace in %q\n", traceFile)
		f, err := os.Create(traceFile)
		if err != nil {
			log.Fatalf("Failed to create go trace output file: %v", err)
		}
		defer func() {
			fmt.Fprintf(os.Stderr, "go trace: go tool trace %s\n", traceFile)
			cerr := f.Close()
			if cerr != nil {
				log.Fatalf("Failed to close go trace output file: %v", cerr)
			}
		}()
		if err := trace.Start(f); err != nil {
			log.Fatalf("Failed to start go trace: %v", err)
		}
		defer trace.Stop()
	}

	return subcommands.Run(getApplication(), nil)
}
