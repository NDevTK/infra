// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//go:build windows

package builtins

import (
	"encoding/binary"
	"os"

	"golang.org/x/sys/windows"
)

const SymlinkCookie = "!<symlink>"

// Create a plain file includes src. This is not the real symlink but accepted
// as a workaround for cygwin when using winsymlinks:sys.
// See: https://cygwin.com/cygwin-ug-net/using-cygwinenv.html
func symlink(src, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	// Symlink magic header
	if _, err := f.Write([]byte(SymlinkCookie)); err != nil {
		return err
	}

	// BOM
	if err := binary.Write(f, binary.LittleEndian, uint16(0xfeff)); err != nil {
		return err
	}

	// Target
	b, err := windows.UTF16FromString(src)
	if err != nil {
		return err
	}
	for _, u := range b {
		if err := binary.Write(f, binary.LittleEndian, u); err != nil {
			return err
		}
	}

	dstW, err := windows.UTF16PtrFromString(dst)
	if err != nil {
		return err
	}
	if err := windows.SetFileAttributes(dstW, windows.FILE_ATTRIBUTE_SYSTEM); err != nil {
		return err
	}
	return nil
}
