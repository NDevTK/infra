// Copyright 2022 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/provenance/client"
	"go.chromium.org/luci/provenance/reporter"
)

const snoopy_addr = "http://localhost:11000"

func main() {
	ctx := context.Background()
	ctx, shutdown := lucictx.TrackSoftDeadline(ctx, 500*time.Millisecond)
	defer shutdown()

	if len(os.Args) < 2 {
		log.Fatalf("not enough arguments passed")
	}
	// Check for the presence of "--" between the wrapper & the cmd
	// bbagent already resolves the cmd to absolute path
	args := os.Args[1:]
	if args[0] != "--" {
		log.Fatalf("there should be  a `--` between the wrapper and the cmd")
	}
	args = args[1:]

	if err := reportPID(ctx, snoopy_addr); err != nil {
		log.Fatalf("failed to report pid to snoopy: %v", err)
	}

	if err := RunInNsjail(ctx, args); err != nil {
		log.Fatalf("running command: %s", err.Error())
	}
}

// reportPID attempts to report pid before calling nsjail.
//
// This is used to track processes in snoopy. Self reported pid inside jail is
// always 1 as process space is different inside jail.
func reportPID(ctx context.Context, snoopy_url string) error {
	reporterClient, err := client.MakeProvenanceClient(ctx, snoopy_url)
	if err != nil {
		log.Fatalf("failed to create snoopy client: %v", err)
		return err
	}

	reporter := &reporter.Report{
		RClient: reporterClient,
	}

	pid := os.Getpid()
	log.Printf("trying to report pid: %d to snoopy", pid)
	if _, err := reporter.ReportPID(ctx, int64(pid)); err != nil {
		log.Fatalf("failed to report pid to snoopy: %v", err)
		return err
	}
	log.Printf("successfully reported pid: %d to snoopy", pid)
	return nil
}
