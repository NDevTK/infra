// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ninja

import (
	"bytes"
	"context"
	"encoding/json"
	"infra/build/siso/build"
	"infra/build/siso/hashfs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuild_Restat(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	exists := func(fname string) error {
		_, err := os.Stat(filepath.Join(dir, "out/siso", fname))
		return err
	}

	hashfsOpts := hashfs.Option{
		StateFile: ".siso_fs_state",
	}

	func() {
		t.Logf("first build")
		setupFiles(t, dir, t.Name(), nil)
		opt, graph, cleanup := setupBuild(ctx, t, dir, hashfsOpts)
		defer cleanup()

		b, err := build.New(ctx, graph, opt)
		if err != nil {
			t.Fatal(err)
		}
		defer b.Close()
		err = b.Build(ctx, "build", "all")
		if err != nil {
			t.Fatalf(`b.Build(ctx, "build", "all")=%v; want nil err`, err)
		}
		if err := exists("foo.out"); err != nil {
			t.Errorf("foo.out doesn't exist: %v", err)
		}
		if err := exists("bar.out"); err != nil {
			t.Errorf("bar.out doesn't exist: %v", err)
		}
	}()

	touch := func(fname string) {
		t.Helper()
		fullname := filepath.Join(dir, fname)
		fi, err := os.Stat(fullname)
		if err != nil {
			err = os.WriteFile(fullname, nil, 0644)
			if err != nil {
				t.Fatal(err)
			}
		}
		for {
			err = os.Chtimes(fullname, time.Now(), time.Now())
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

	func() {
		t.Logf("second build. touch base/foo.in")
		touch("base/foo.in")
		opt, graph, cleanup := setupBuild(ctx, t, dir, hashfsOpts)
		defer cleanup()
		var metricsBuffer bytes.Buffer
		opt.MetricsJSONWriter = &metricsBuffer

		b, err := build.New(ctx, graph, opt)
		if err != nil {
			t.Fatal(err)
		}
		defer b.Close()

		err = b.Build(ctx, "build", "all")
		if err != nil {
			t.Fatalf(`b.Build(ctx, "build", "all")=%v; want nil err`, err)
		}
		stat := b.Stats()
		if stat.Skipped != 2 {
			t.Errorf("Skipped=%d; want 2", stat.Skipped)
		}
		dec := json.NewDecoder(bytes.NewReader(metricsBuffer.Bytes()))
		for dec.More() {
			var m build.StepMetric
			err := dec.Decode(&m)
			if err != nil {
				t.Errorf("decode %v", err)
			}
			if m.StepID == "" {
				continue
			}
			switch filepath.Base(m.Output) {
			case "foo.out":
				if m.Err {
					t.Errorf("%s err=%t; want false", m.Output, m.Err)
				}
			default:
				t.Errorf("unexpected output %q: %#v", m.Output, m)
			}
		}
	}()

	update := func(fname string) {
		t.Helper()
		fullname := filepath.Join(dir, fname)
		buf, err := os.ReadFile(fullname)
		if err != nil {
			t.Fatal(err)
		}
		buf = append(buf, []byte("!!!")...)
		for {
			fi, err := os.Stat(fullname)
			if err != nil {
				t.Fatal(err)
			}
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

	func() {
		t.Logf("third build, update base/foo.in")
		update("base/foo.in")
		opt, graph, cleanup := setupBuild(ctx, t, dir, hashfsOpts)
		defer cleanup()
		var metricsBuffer bytes.Buffer
		opt.MetricsJSONWriter = &metricsBuffer

		b, err := build.New(ctx, graph, opt)
		if err != nil {
			t.Fatal(err)
		}
		defer b.Close()

		err = b.Build(ctx, "build", "all")
		if err != nil {
			t.Fatalf(`b.Build(ctx, "build", "all")=%v; want nil err`, err)
		}
		stat := b.Stats()
		if stat.Skipped != 1 {
			t.Errorf("Skipped=%d; want 1", stat.Skipped)
		}
		dec := json.NewDecoder(bytes.NewReader(metricsBuffer.Bytes()))
		for dec.More() {
			var m build.StepMetric
			err := dec.Decode(&m)
			if err != nil {
				t.Errorf("decode %v", err)
			}
			if m.StepID == "" {
				continue
			}
			switch filepath.Base(m.Output) {
			case "foo.out", "bar.out":
				if m.Err {
					t.Errorf("%s err=%t; want false", m.Output, m.Err)
				}
			default:
				t.Errorf("unexpected output %q", m.Output)
			}
		}

	}()
}
