// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package testutils

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	"go.chromium.org/luci/common/errors"

	"github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
	"github.com/spf13/afero/mem"
)

var _ afero.Symlinker = AferoMemMapFs{}

// Add Symlinker to afero.MemMapFS. This is useful for testing.
type AferoMemMapFs struct {
	*afero.MemMapFs
}

func NewAferoMemMapFs() afero.Fs {
	return AferoMemMapFs{
		MemMapFs: &afero.MemMapFs{},
	}
}

func (m AferoMemMapFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	f, err := m.Open(name)
	if err != nil {
		return nil, false, err
	}
	fi := mem.GetFileInfo(f.(*mem.File).Data())
	return fi, fi.Mode().Type() == fs.ModeSymlink, nil
}

func (m AferoMemMapFs) Stat(name string) (os.FileInfo, error) {
	for i := 0; i < 1024; i++ {
		f, err := m.Open(name)
		if err != nil {
			return nil, err
		}
		fi := mem.GetFileInfo(f.(*mem.File).Data())
		if fi.Mode().Type() != fs.ModeSymlink {
			return fi, nil
		}
		if name, err = m.ReadlinkIfPossible(name); err != nil {
			return nil, err
		}
	}
	return nil, afero.ErrTooLarge
}

func (m AferoMemMapFs) ReadlinkIfPossible(name string) (string, error) {
	f, err := m.Open(name)
	if err != nil {
		return "", err
	}
	fi := mem.GetFileInfo(f.(*mem.File).Data())
	if fi.Mode().Type() != os.ModeSymlink {
		return "", fs.ErrInvalid
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (m AferoMemMapFs) SymlinkIfPossible(oldname, newname string) error {
	f, err := m.Create(newname)
	if err != nil {
		return err
	}
	mem.SetMode(f.(*mem.File).Data(), os.ModeSymlink)
	if _, err := f.WriteString(oldname); err != nil {
		return err
	}
	return nil
}

// Add ReadLinkFS to afero.IOFS. This is useful for testing.
type AferoIOFS struct {
	*afero.IOFS
}

func NewAferoIOFS(fs afero.Fs) AferoIOFS {
	return AferoIOFS{
		IOFS: &afero.IOFS{Fs: fs},
	}
}

func (iofs AferoIOFS) ReadLink(name string) (string, error) {
	return iofs.IOFS.Fs.(afero.LinkReader).ReadlinkIfPossible(name)
}

// copyFS is copied from actions/copy.go
func CopyFS(src fs.FS, dstFS afero.Fs) {
	convey.So(fs.WalkDir(src, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		log.Printf("copying %s\n", name)

		var cerr error
		err = func() error {
			switch {
			case d.Type() == fs.ModeDir:
				if err := dstFS.MkdirAll(name, fs.ModePerm); err != nil {
					return fmt.Errorf("failed to create dir: %s: %w", name, err)
				}
				return nil
			case d.Type() == fs.ModeSymlink:
				fallthrough
			default: // Regular File
				srcFile, err := src.Open(name)
				if err != nil {
					return fmt.Errorf("failed to open src file: %s: %w", name, err)
				}
				defer func() { cerr = errors.Join(cerr, srcFile.Close()) }()

				info, err := fs.Stat(src, name)
				if err != nil {
					return fmt.Errorf("failed to get file mode: %s: %w", name, err)
				}

				if err := createFile(name, srcFile, info.Mode(), dstFS); err != nil {
					return fmt.Errorf("failed to create file: %s: %w", name, err)
				}
			}
			return nil
		}()

		return errors.Join(err, cerr)
	}), convey.ShouldBeNil)
}

// createFile is copied from actions/copy.go
func createFile(dst string, r io.Reader, m fs.FileMode, dstFS afero.Fs) error {
	f, err := dstFS.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	if err := dstFS.Chmod(dst, m); err != nil {
		return err
	}
	return nil
}
