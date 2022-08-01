// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"embed"
	"fmt"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
)

var (
	//go:embed setup_linux.py
	setupLinuxFiles embed.FS
	setupLinux      = &builtins.CopyFiles{
		Name:  "setup_linux",
		Files: setupLinuxFiles,
	}
)

func (g *Generator) getSource() (cipkg.Generator, error) {
	switch s := g.Source.(type) {
	case *SourceGit:
		panic("unimplemented")
	case *SourceURL:
		return &builtins.FetchURL{
			Name:          fmt.Sprintf("%s_source", g.Name),
			URL:           s.URL,
			Filename:      s.Filename,
			HashAlgorithm: s.HashAlgorithm,
			HashString:    s.HashString,
		}, nil
	default:
		return nil, fmt.Errorf("unknown source type %#v:", s)
	}
}

// pre_hook is defined as a function in python to create a closure because
// execute_cmd and activate_pkg share some variables.
const preHook = `
def pre_hook(exe):
  dependencies = []

  def execute_cmd(exe):
    ctx = exe.current_context
    cwd = os.getcwd()
    out = exe.env['out']

    volumes = [
        '--volume', f'{cwd}:{cwd}',
        '--volume', f'{out}:{out}',
    ]
    for dep in dependencies:
      volumes.extend(('--volume', f'{dep}:{dep}'))
  
    docker = [
        'docker', 'run', '--rm',
        '--workdir', cwd,
        '--user', f'{os.getuid()}:{os.getgid()}',
    ]

    impage = [
        'gcr.io/chromium-container-registry/infra-dockerbuild/manylinux-x64-py3:v1.4.18',
    ]

    subprocess.check_call(docker + volumes + impage + ctx.args, env=exe.env)
    return True

  def activate_pkg(exe):
    ctx = exe.current_context
    dependencies.append(str(ctx.pkg))
    return True

  exe.add_hook('executeCmd', execute_cmd)
  exe.add_hook('activatePkg', activate_pkg)


pre_hook(exe)
`

func (g *Generator) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	src, err := g.getSource()
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
	}

	base := &utilities.BaseGenerator{
		Name:    g.Name,
		Builder: "{{.stdenv_python3}}/bin/python3",
		Args:    []string{"-I", "-B", "{{.setup_linux}}/setup_linux.py", "{{.stdenv}}"},
		Env: append([]string{
			"buildFlags=",
			"installFlags=",
			fmt.Sprintf("srcs={{.%s_source}}", g.Name),
		}, g.Env...),
		Dependencies: append([]utilities.BaseDependency{
			{Type: cipkg.DepsBuildHost, Generator: src},
			{Type: cipkg.DepsBuildHost, Generator: common.Stdenv},
			{Type: cipkg.DepsBuildHost, Generator: common.PosixUtils},
			{Type: cipkg.DepsBuildHost, Generator: common.Docker},
			{Type: cipkg.DepsBuildHost, Generator: common.Git},
			{Type: cipkg.DepsBuildHost, Generator: common.Python3},
			{Type: cipkg.DepsBuildHost, Generator: setupLinux},
		}, g.Dependencies...),
	}
	return base.Generate(ctx)
}
