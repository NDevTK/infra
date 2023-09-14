// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ninja

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"infra/build/siso/build"
	"infra/build/siso/hashfs"
)

// Test rebuild build.ninja (gn gen) behavior.
func TestBuild_GN(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	run := func(args ...string) error {
		cmd := exec.Command("python3", args...)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	sisoNinja := func(s string) (build.Stats, []build.StepMetric, error) {
		t.Logf("build - %s", s)
		t.Helper()
		opt, graph, cleanup := setupBuild(ctx, t, dir, hashfs.Option{
			StateFile: ".siso_fs_state",
		})
		defer cleanup()

		mfopt := opt
		mfopt.Clobber = false
		mfopt.RebuildManifest = "build.ninja"
		mfb, err := build.New(ctx, graph, mfopt)
		if err != nil {
			t.Fatal(err)
		}
		err = mfb.Build(ctx, "rebuild manifest", "build.ninja")
		mfb.Close()
		stats := mfb.Stats()
		if errors.Is(err, build.ErrManifestModified) {
			return stats, nil, err
		}
		if err != nil {
			t.Fatal(err)
		}

		var metricsBuffer bytes.Buffer
		opt.MetricsJSONWriter = &metricsBuffer

		b, err := build.New(ctx, graph, opt)
		if err != nil {
			t.Fatal(err)
		}
		defer b.Close()
		err = b.Build(ctx, "build", "all")
		stats = b.Stats()
		var metrics []build.StepMetric
		dec := json.NewDecoder(bytes.NewReader(metricsBuffer.Bytes()))
		for dec.More() {
			var m build.StepMetric
			derr := dec.Decode(&m)
			if derr != nil {
				t.Errorf("decode %v", derr)
			}
			metrics = append(metrics, m)
		}
		return stats, metrics, err
	}

	setupFiles(t, dir, t.Name(), nil)
	err := run("buildtools/platform/gn.py", "gen", "out/siso")
	if err != nil {
		t.Fatalf("gn gen failed: %v", err)
	}

	stats, _, err := sisoNinja("first")
	if err != nil {
		t.Fatalf("first build=%v; want nil err", err)
	}
	if stats.Total != 4 {
		t.Errorf("first build Total=%d want=4", stats.Total)
	}
	if stats.Done != 4 {
		t.Errorf("first build Done=%d want=4", stats.Done)
	}
	if stats.Skipped != 1 {
		t.Errorf("first build Skipped=%d want=1", stats.Skipped)
	}
	stats, _, err = sisoNinja("null")
	if err != nil {
		t.Fatalf("null build=%v; want nil err", err)
	}
	if stats.Total != 4 {
		t.Errorf("null build Total=%d want=4", stats.Total)
	}
	if stats.Done != 4 {
		t.Errorf("null build Done=%d want=4", stats.Done)
	}
	if stats.Skipped != 4 {
		t.Errorf("null build Skipped=%d want=4", stats.Skipped)
	}

	update := func(fname string) {
		t.Helper()
		fullname := filepath.Join(dir, fname)
		fi, err := os.Stat(fullname)
		if err != nil {
			t.Fatal(err)
		}
		buf, err := os.ReadFile(fullname)
		if err != nil {
			t.Fatal(err)
		}
		buf = append(buf, []byte("!!!")...)
		for {
			err = os.WriteFile(fullname, buf, 0644)
			if err != nil {
				t.Fatal(err)
			}
			nfi, err := os.Stat(fullname)
			if err != nil {
				t.Fatal(err)
			}
			if fi.ModTime().Equal(nfi.ModTime()) {
				time.Sleep(1 * time.Millisecond)
				continue
			}
			return
		}
	}

	update("BUILD.gn")

	stats, _, err = sisoNinja("incremental-regen")
	if !errors.Is(err, build.ErrManifestModified) {
		t.Fatalf("regen build=%v; want %v", err, build.ErrManifestModified)
	}

	stats, _, err = sisoNinja("incremental-after-regen")
	if err != nil {
		t.Fatalf("incremental build after regen %v, want nil err", err)
	}
	if stats.Total != 4 {
		t.Errorf("incremental build Total=%d want=4", stats.Total)
	}
	if stats.Done != 4 {
		t.Errorf("incremental build Done=%d want=4", stats.Done)
	}
	if stats.Skipped != 4 {
		t.Errorf("incremental build Skipped=%d want=4", stats.Skipped)
	}

	err = run("buildtools/platform/gn.py", "clean", "out/siso")
	if err != nil {
		t.Fatalf("gn clean failed: %v", err)
	}

	stats, _, err = sisoNinja("clean-regen")
	if !errors.Is(err, build.ErrManifestModified) {
		t.Fatalf("clean build=%v; want %v", err, build.ErrManifestModified)
	}

	stats, _, err = sisoNinja("clean-after-regen")
	if err != nil {
		t.Fatalf("clean build after regen %v, want nil err", err)
	}
	if stats.Total != 4 {
		t.Errorf("clean build Total=%d want=4", stats.Total)
	}
	if stats.Done != 4 {
		t.Errorf("clean build Done=%d want=4", stats.Done)
	}
	if stats.Skipped != 1 {
		t.Errorf("clean build Skipped=%d want=4", stats.Skipped)
	}
}
