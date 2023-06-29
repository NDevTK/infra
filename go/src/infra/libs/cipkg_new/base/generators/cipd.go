// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package generators

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"

	"infra/libs/cipkg_new/base/actions"
	"infra/libs/cipkg_new/core"

	"go.chromium.org/luci/cipd/client/cipd"
	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/cipd/client/cipd/template"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/system/environ"
)

// CIPDExport is used for downloading CIPD packages. It behaves similar to
// `cipd export` for the provided ensure file and use ${out} as the cipd
// root path.
// TODO(crbug/1323147): Replace direct call cipd binary with cipd sdk when it's
// available.
type CIPDExport struct {
	Metadata *core.Action_Metadata

	Ensure   ensure.File
	Expander template.Expander

	ConfigFile          string
	CacheDir            string
	HTTPUserAgentPrefix string
	MaxThreads          int
	ParallelDownloads   int
	ServiceURL          string
}

func (c *CIPDExport) Generate(ctx context.Context, plats Platforms) (*core.Action, error) {
	// Expand template based on ctx.Platform.Host before pass it to cipd client
	// for cross-compile
	expander := c.Expander
	if expander == nil {
		expander = template.Platform{
			OS:   cipdOS(plats.Host.OS()),
			Arch: cipdArch(plats.Host.Arch()),
		}.Expander()
	}

	ef, err := expandEnsureFile(&c.Ensure, expander)
	if err != nil {
		return nil, fmt.Errorf("failed to expand ensure file: %v: %w", c.Ensure, err)
	}

	var w strings.Builder
	if err := ef.Serialize(&w); err != nil {
		return nil, fmt.Errorf("failed to encode ensure file: %v: %w", ef, err)
	}

	env := environ.New(nil)
	addEnv := func(k string, v any) {
		if vv := reflect.ValueOf(v); !vv.IsValid() || vv.IsZero() {
			return
		}
		env.Set(k, fmt.Sprintf("%v", v))
	}

	addEnv(cipd.EnvConfigFile, c.ConfigFile)
	addEnv(cipd.EnvCacheDir, c.CacheDir)
	addEnv(cipd.EnvHTTPUserAgentPrefix, c.HTTPUserAgentPrefix)
	addEnv(cipd.EnvMaxThreads, c.MaxThreads)
	addEnv(cipd.EnvParallelDownloads, c.ParallelDownloads)
	addEnv(cipd.EnvCIPDServiceURL, c.ServiceURL)

	return &core.Action{
		Metadata: c.Metadata,
		Deps:     []*core.Action_Dependency{actions.ReexecDependency()},
		Spec: &core.Action_Cipd{
			Cipd: &core.ActionCIPDExport{
				EnsureFile: w.String(),
				Env:        env.Sorted(),
			},
		},
	}, nil
}

func cloneEnsureFile(ef *ensure.File) (*ensure.File, error) {
	var s bytes.Buffer
	if err := ef.Serialize(&s); err != nil {
		return nil, err
	}
	return ensure.ParseFile(&s)
}

func expandEnsureFile(ef *ensure.File, expander template.Expander) (*ensure.File, error) {
	ef, err := cloneEnsureFile(ef)
	if err != nil {
		return nil, err
	}

	for dir, slice := range ef.PackagesBySubdir {
		var s ensure.PackageSlice
		for _, p := range slice {
			pkg, err := p.Expand(expander)
			switch err {
			case template.ErrSkipTemplate:
				continue
			case nil:
			default:
				return nil, errors.Annotate(err, "expanding %#v", pkg).Err()
			}
			p.PackageTemplate = pkg
			s = append(s, p)
		}
		ef.PackagesBySubdir[dir] = s
	}
	return ef, nil
}

// TODO(fancl): move these to go.chromium.org/luci/cipd as utilities.
func cipdOS(os string) string {
	if os == "darwin" {
		return "mac"
	}
	return os
}

func cipdArch(arch string) string {
	if arch == "arm" {
		return "armv6l"
	}
	return arch
}
