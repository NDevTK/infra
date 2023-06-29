// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package actions

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"

	"infra/libs/cipkg_new/core"

	"github.com/spf13/afero"
	"go.chromium.org/luci/common/system/environ"
)

// ActionFilesCopyTransformer is the default transformer for copy action.
func ActionFilesCopyTransformer(a *core.ActionFilesCopy, deps []PackageDependency) (*core.Derivation, error) {
	drv, err := ReexecDerivation(a, deps, false)
	if err != nil {
		return nil, err
	}

	for _, d := range deps {
		name := d.Package.Action.Metadata.Name
		outDir := d.Package.Handler.OutputDirectory()
		drv.Env = append(drv.Env, fmt.Sprintf("%s=%s", name, outDir))
	}
	return drv, nil
}

var defaultFilesCopyExecutor = FilesCopyExecutor{}

// RegisterEmbed regists the embedded fs with ref. It can be retrieved by copy
// actions using embed source. Embedded fs need to be registered in init() for
// re-exec executor.
func RegisterEmbed(ref string, e embed.FS) {
	defaultFilesCopyExecutor.StoreEmbed(ref, e)
}

// ActionCIPDExportExecutor is the default executor for cipd action.
// All embed.FS must be registered in init() so they are available when being
// executed from reexec,
type FilesCopyExecutor struct {
	embeds sync.Map
}

// StoreEmbed registers an embedded fs for copy executor. This need to be
// called before main function otherwise executor may not be able to load the
// fs. It's caller's responsibility to ensure ref is unique.
// TODO(fancl): record conflicts.
func (f *FilesCopyExecutor) StoreEmbed(ref string, e embed.FS) {
	f.embeds.LoadOrStore(ref, e)
}

// LoadEmbed load a registered embedded fs.
func (f *FilesCopyExecutor) LoadEmbed(ref string) (embed.FS, bool) {
	e, ok := f.embeds.Load(ref)
	if !ok {
		return embed.FS{}, false
	}
	return e.(embed.FS), true
}

func (f *FilesCopyExecutor) Execute(ctx context.Context, a *core.ActionFilesCopy, dstFS afero.Fs) error {
	for dst, srcFile := range a.Files {
		if err := dstFS.MkdirAll(path.Dir(dst), fs.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %s: %w", path.Base(dst), err)
		}
		m := fs.FileMode(srcFile.Mode)

		switch c := srcFile.Content.(type) {
		case *core.ActionFilesCopy_Source_Raw:
			if err := copyRaw(c.Raw, dst, m, dstFS); err != nil {
				return err
			}
		case *core.ActionFilesCopy_Source_Local_:
			if err := copyLocal(c.Local, m, dst, dstFS); err != nil {
				return err
			}
		case *core.ActionFilesCopy_Source_Embed_:
			if err := f.copyEmbed(c.Embed, m, dst, dstFS); err != nil {
				return err
			}
		case *core.ActionFilesCopy_Source_Output_:
			if err := copyOutput(c.Output, environ.FromCtx(ctx), m, dst, dstFS); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown file type for %s: %s", dst, m.Type())
		}
	}
	return nil
}

func copyRaw(raw []byte, dst string, m fs.FileMode, dstFS afero.Fs) error {
	switch m.Type() {
	case fs.ModeSymlink:
		return fmt.Errorf("symlink is not supported for the source type")
	case fs.ModeDir:
		if err := dstFS.MkdirAll(dst, m); err != nil {
			return fmt.Errorf("failed to create directory: %s: %w", path.Base(dst), err)
		}
		return nil
	case 0: // Regular File
		if err := createFile(dst, bytes.NewReader(raw), m, dstFS); err != nil {
			return fmt.Errorf("failed to create file: %s: %w", dst, err)
		}
		return nil
	default:
		return fmt.Errorf("unknown file type for %s: %s", dst, m.Type())
	}
}

func copyLocal(s *core.ActionFilesCopy_Source_Local, m fs.FileMode, dst string, dstFS afero.Fs) error {
	src := s.Path
	switch m.Type() {
	case fs.ModeSymlink:
		if s.FollowSymlinks {
			return fmt.Errorf("invalid file spec: followSymlinks can't be used with symlink dst: %s", dst)
		}
		dstFS, ok := dstFS.(afero.Symlinker)
		if !ok {
			return fmt.Errorf("symlink not supported on the destination filesystem: %s", dst)
		}
		if err := dstFS.SymlinkIfPossible(src, dst); err != nil {
			return fmt.Errorf("failed to create symlink: %s -> %s: %w", dst, src, err)
		}
		return nil
	case fs.ModeDir:
		return copyFS(os.DirFS(src), s.FollowSymlinks, NewBasePathFs(dstFS, dst))
	case 0: // Regular File
		f, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("failed to open source file for %s: %s: %w", dst, src, err)
		}
		defer f.Close()
		if err := createFile(dst, f, m, dstFS); err != nil {
			return fmt.Errorf("failed to create file for %s: %s: %w", dst, src, err)
		}
		return nil
	default:
		return fmt.Errorf("unknown file type for %s: %s", dst, m.Type())
	}
}

func (f *FilesCopyExecutor) copyEmbed(s *core.ActionFilesCopy_Source_Embed, m fs.FileMode, dst string, dstFS afero.Fs) error {
	e, ok := f.LoadEmbed(s.Ref)
	if !ok {
		return fmt.Errorf("failed to load embedded fs for %s: %s", dst, s)
	}

	switch m.Type() {
	case fs.ModeSymlink:
		return fmt.Errorf("symlink not supported for the source type")
	case fs.ModeDir:
		path := s.Path
		if path == "" {
			path = "."
		}
		src, err := fs.Sub(e, path)
		if err != nil {
			return fmt.Errorf("failed to load subdir fs for %s: %s: %w", dst, s, err)
		}
		return copyFS(src, true, NewBasePathFs(dstFS, dst))
	case 0: // Regular File
		f, err := e.Open(s.Path)
		if err != nil {
			return fmt.Errorf("failed to open source file for %s: %s: %w", dst, s, err)
		}
		defer f.Close()
		if err := createFile(dst, f, m, dstFS); err != nil {
			return fmt.Errorf("failed to create file for %s: %s: %w", dst, s, err)
		}
		return nil
	default:
		return fmt.Errorf("unknown file type for %s: %s", dst, m.Type())
	}
}

func copyOutput(s *core.ActionFilesCopy_Source_Output, env environ.Env, m fs.FileMode, dst string, dstFS afero.Fs) error {
	out := env.Get(s.Name)
	if out == "" {
		return fmt.Errorf("output not found: %s: %s", dst, s)
	}
	return copyLocal(&core.ActionFilesCopy_Source_Local{
		Path:           filepath.Join(out, s.Path),
		FollowSymlinks: false,
	}, m, dst, dstFS)
}

func copyFS(src fs.FS, followSymlinks bool, dstFS afero.Fs) error {
	return fs.WalkDir(src, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		log.Printf("copying %s\n", name)

		var cerr error
		err = func() error {
			switch {
			case d.Type() == fs.ModeSymlink && !followSymlinks:
				src, ok := src.(ReadLinkFS)
				if !ok {
					return fmt.Errorf("readlink not supported on the source filesystem: %s", name)
				}
				target, err := src.ReadLink(name)
				if err != nil {
					return fmt.Errorf("failed to readlink: %s: %w", name, err)
				}
				if filepath.IsAbs(target) {
					return fmt.Errorf("absolute symlink target not supported: %s", name)
				}

				dstFS, ok := dstFS.(afero.Symlinker)
				if !ok {
					return fmt.Errorf("symlink not supported on the destination filesystem: %s", name)
				}
				if err := dstFS.SymlinkIfPossible(target, name); err != nil {
					return fmt.Errorf("failed to create symlink: %s -> %s: %w", name, target, err)
				}
			case d.Type() == fs.ModeDir:
				if err := dstFS.MkdirAll(name, fs.ModePerm); err != nil {
					return fmt.Errorf("failed to create dir: %s: %w", name, err)
				}
				return nil
			default: // Regular File or following symlinks
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
	})
}

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
