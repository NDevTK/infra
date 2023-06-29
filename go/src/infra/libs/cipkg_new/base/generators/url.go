// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package generators

import (
	"context"
	"fmt"
	"io/fs"

	"infra/libs/cipkg_new/base/actions"
	"infra/libs/cipkg_new/core"
)

type FetchURL struct {
	URL           string
	Mode          fs.FileMode
	HashAlgorithm core.HashAlgorithm
	HashValue     string
}

// FetchURLs downloads files from servers based on the path-url pairs of URLs.
type FetchURLs struct {
	Metadata *core.Action_Metadata
	URLs     map[string]FetchURL
}

func (f *FetchURLs) Generate(ctx context.Context, plats Platforms) (*core.Action, error) {
	deps := []*core.Action_Dependency{actions.ReexecDependency()}

	// Generate separate action for every url.
	files := make(map[string]*core.ActionFilesCopy_Source)
	for k, v := range f.URLs {
		spec := &core.ActionURLFetch{
			Url:           v.URL,
			HashAlgorithm: v.HashAlgorithm,
			HashValue:     v.HashValue,
		}

		// 32^6 = 2^30 should be good enough.
		id, err := core.StableID(spec, 6)
		if err != nil {
			return nil, err
		}
		name := fmt.Sprintf("%s_%s", f.Metadata.Name, id)

		deps = append(deps, &core.Action_Dependency{
			Action: &core.Action{
				Metadata: &core.Action_Metadata{Name: name},
				Deps:     []*core.Action_Dependency{actions.ReexecDependency()},
				Spec:     &core.Action_Url{Url: spec},
			},
			Runtime: false,
		})

		m := v.Mode
		if m == 0 {
			m = 0o666
		}
		files[k] = &core.ActionFilesCopy_Source{
			Content: &core.ActionFilesCopy_Source_Output_{
				Output: &core.ActionFilesCopy_Source_Output{Name: name, Path: "file"},
			},
			Mode: uint32(m),
		}
	}

	// TODO: Sort deps
	return &core.Action{
		Metadata: f.Metadata,
		Deps:     deps,
		Spec: &core.Action_Copy{
			Copy: &core.ActionFilesCopy{
				Files: files,
			},
		},
	}, nil
}
