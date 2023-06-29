// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package actions

import (
	"io/fs"
	"os"

	"github.com/spf13/afero"
)

// TODO(fancl): Consider remove the usage of afero & io/fs for tests. Temporary
// directory may be fine for us.

// ReadLinkFS is the interface implemented by a file system that supports
// symbolic links.
// TODO(fancl): Replace it with https://github.com/golang/go/issues/49580
type ReadLinkFS interface {
	fs.FS

	// ReadLink returns the destination of the named symbolic link.
	// Link destinations will always be slash-separated paths.
	// NOTE: Although in the standard library the link destination is guaranteed
	// to be a path inside FS. We may return host destination for our use cases.
	ReadLink(name string) (string, error)
}

// BasePathFs is almost same as afero.BasePathFs except SymlinkIfPossible won't
// transform the oldname based on the base path. It makes BasePathFs consistent
// because ReadlinkIfPossible won't transform the return address.
type BasePathFs struct {
	source afero.Fs
	*afero.BasePathFs
}

func NewBasePathFs(source afero.Fs, path string) afero.Fs {
	return &BasePathFs{
		source:     source,
		BasePathFs: afero.NewBasePathFs(source, path).(*afero.BasePathFs),
	}
}

func (b *BasePathFs) SymlinkIfPossible(oldname, newname string) error {
	newname, err := b.RealPath(newname)
	if err != nil {
		return &os.LinkError{Op: "symlink", Old: oldname, New: newname, Err: err}
	}
	if linker, ok := b.source.(afero.Linker); ok {
		return linker.SymlinkIfPossible(oldname, newname)
	}
	return &os.LinkError{Op: "symlink", Old: oldname, New: newname, Err: afero.ErrNoSymlink}
}
