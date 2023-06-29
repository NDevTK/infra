// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package generators

import (
	"context"
	"crypto"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"

	"infra/libs/cipkg_new/base/actions"
	"infra/libs/cipkg_new/core"
)

type EmbeddedFiles struct {
	Metadata *core.Action_Metadata

	ref string
	dir string
	efs embed.FS

	modeOverride func(info fs.FileInfo) (fs.FileMode, error)
}

// InitEmbeddedFS regists the embed.FS to copy executor and returns the
// corresponding generator.
func InitEmbeddedFS(metadata *core.Action_Metadata, e embed.FS) *EmbeddedFiles {
	h := crypto.SHA256.New()
	if err := getHashFromFS(e, h); err != nil {
		// Embedded files are frozen after build. Panic since this is more like a
		// build failure.
		panic(err)
	}
	ref := fmt.Sprintf("%x", h.Sum(nil))
	actions.RegisterEmbed(ref, e)
	return &EmbeddedFiles{
		Metadata: metadata,

		ref: ref,
		dir: ".",
		efs: e,
	}
}

func (e *EmbeddedFiles) Generate(ctx context.Context, plats Platforms) (*core.Action, error) {
	efs, err := fs.Sub(e.efs, e.dir)
	if err != nil {
		return nil, err
	}

	files := make(map[string]*core.ActionFilesCopy_Source)
	if err := fs.WalkDir(efs, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if name == "." {
			return fmt.Errorf("root dir can't be file.")
		}

		mode := fs.FileMode(0o666)
		if e.modeOverride != nil {
			info, err := d.Info()
			if err != nil {
				return err
			}
			if mode, err = e.modeOverride(info); err != nil {
				return err
			}
		}

		files[filepath.FromSlash(name)] = &core.ActionFilesCopy_Source{
			Content: &core.ActionFilesCopy_Source_Embed_{
				Embed: &core.ActionFilesCopy_Source_Embed{Ref: e.ref, Path: path.Join(e.dir, name)},
			},
			Mode: uint32(mode),
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &core.Action{
		Metadata: e.Metadata,
		Deps:     []*core.Action_Dependency{actions.ReexecDependency()},
		Spec: &core.Action_Copy{
			Copy: &core.ActionFilesCopy{Files: files},
		},
	}, nil
}

// SubDir returns a generator copies files in the sub directory of the source.
func (e *EmbeddedFiles) SubDir(dir string) *EmbeddedFiles {
	ret := *e
	ret.dir = path.Join(e.dir, dir)
	return &ret
}

// WithModeOverride overrides file modes while copying.
func (e *EmbeddedFiles) WithModeOverride(f func(info fs.FileInfo) (fs.FileMode, error)) *EmbeddedFiles {
	ret := *e
	ret.modeOverride = f
	return &ret
}
