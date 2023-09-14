// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	cryptorand "crypto/rand"
	"flag"
	"os/exec"
	"path/filepath"
	"testing"

	"infra/experimental/golangbuild/golangbuildpb"
)

func TestGoDistListRebalance(t *testing.T) {
	if testing.Short() || !testing.Verbose() || flag.Lookup("test.run").Value.String() != "^TestGoDistListRebalance$" {
		t.Skip("not running test used for manual shard rebalancing if not explicitly requested with go test -v -run=^TestGoDistListRebalance$")
	}

	gorootBinGo, err := exec.LookPath("go")
	if err != nil {
		t.Skip("did not find 'go' in PATH:", err)
	}
	mustGoDistList := func(shard testShard) []Port {
		t.Helper()
		ports, err := goDistList(context.Background(), &buildSpec{
			inputs: &golangbuildpb.Inputs{},
			goroot: filepath.Join(filepath.Dir(gorootBinGo), ".."),
		}, shard)
		if err != nil {
			t.Fatal("goDistList:", err)
		}
		return ports
	}

	const (
		nShards = 12
		goal    = 3 // The target maximum per shard to look for.
	)
	var goalMet = true
	for shardID := uint32(0); shardID < nShards; shardID++ {
		ports := mustGoDistList(testShard{shardID, nShards})
		goalMet = goalMet && len(ports) <= goal
		t.Logf("shard %d of %d: %d ports: %v", shardID+1, nShards, len(ports), ports)
	}
	totalPorts := mustGoDistList(noSharding)
	t.Logf("a total of %d ports to rebalance", len(totalPorts))
	if goal*nShards < len(totalPorts) {
		t.Fatalf("impossible goal: %d < %d", goal*nShards, len(totalPorts))
	} else if goalMet {
		t.Logf("(goal already met)\n\n")
	}

	goDistListRebalance[nShards] = make([]byte, 32)
Outer:
	for i := 0; ; i++ {
		if i%1000 == 0 {
			t.Logf("attempt %d to find a good rebalance", i)
		}
		_, err := cryptorand.Read(goDistListRebalance[nShards])
		if err != nil {
			t.Fatal(err)
		}

		for shardID := uint32(0); shardID < nShards; shardID++ {
			if len(mustGoDistList(testShard{shardID, nShards})) > goal {
				continue Outer
			}
		}
		t.Logf("found a good rebalance: %#v", goDistListRebalance[nShards])
		break
	}
}
