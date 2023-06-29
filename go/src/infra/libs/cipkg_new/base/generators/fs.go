// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package generators

import (
	"archive/tar"
	"fmt"
	"hash"
	"io"
	"io/fs"
)

func getHashFromFS(src fs.FS, h hash.Hash) error {
	// Tar is used for calculating hash from files - including metadata - in a
	// simple way.
	tw := tar.NewWriter(h)
	defer tw.Close()

	return fs.WalkDir(src, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		var info fs.FileInfo
		// Resolve symlink
		if d.Type() == fs.ModeSymlink {
			f, err := src.Open(name)
			if err != nil {
				return fmt.Errorf("failed to open file: %s: %w", name, err)
			}
			defer f.Close()
			if info, err = f.Stat(); err != nil {
				return fmt.Errorf("failed to stat file: %s: %w", name, err)
			}
		} else {
			if info, err = fs.Stat(src, name); err != nil {
				return fmt.Errorf("failed to stat file: %s: %w", name, err)
			}
		}

		switch info.Mode().Type() {
		case fs.ModeSymlink:
			return fmt.Errorf("unexpected symlink: %s", name)
		case fs.ModeDir:
			if err := tw.WriteHeader(&tar.Header{
				Typeflag: tar.TypeDir,
				Name:     name,
				Mode:     int64(info.Mode()),
			}); err != nil {
				return fmt.Errorf("failed to write header: %s: %w", name, err)
			}
		default: // Regular File
			if err := tw.WriteHeader(&tar.Header{
				Typeflag: tar.TypeReg,
				Name:     name,
				Mode:     int64(info.Mode()),
				Size:     info.Size(),
			}); err != nil {
				return fmt.Errorf("failed to write header: %s: %w", name, err)
			}
			f, err := src.Open(name)
			if err != nil {
				return fmt.Errorf("failed to open file: %s: %w", name, err)
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return fmt.Errorf("failed to write file: %s: %w", name, err)
			}
		}
		return nil
	})
}
