# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Setup script with linux specified hooks."""
# pylint: disable=global-at-module-level
# pylint: disable=undefined-variable

import os
import subprocess
import sys

# Insert the path to stdenv for python searching setup module.
sys.path.insert(0, sys.argv[1])

import setup


def main() -> None:
  dependencies = []

  def execute_cmd(exe) -> bool:
    ctx = exe.current_context
    cwd = os.getcwd()
    out = exe.env['out']
    tmp = exe.env['buildTemp']

    volumes = [
        '--volume', f'{tmp}:{tmp}',
        '--volume', f'{out}:{out}',
    ]
    for dep in dependencies:
      volumes.extend(('--volume', f'{dep}:{dep}'))

    docker = [
        'docker', 'run', '--rm',
        '--workdir', cwd,
        '--user', f'{os.getuid()}:{os.getgid()}',
    ]

    env = []
    for k, v in exe.env.items():
      if k not in {'PATH'}:
        env.extend(('--env', f'{k}={v}'))
    # force override LDFLAGS even it's not set. This is because dockcross by
    # default set it to '-L/usr/cross/lib', which may override the library
    # path passed to the configure.
    if 'LDFLAGS' not in exe.env:
      env.extend(('--env', 'LDFLAGS='))

    impage = [
        exe.env['dockerImage'],
    ]

    subprocess.check_call(docker + volumes + env + impage + ctx.args)
    return True

  def activate_pkg(exe) -> bool:
    ctx = exe.current_context
    dependencies.append(str(ctx.pkg))
    return True

  exe = setup.Execution()
  exe.add_hook('executeCmd', execute_cmd)
  exe.add_hook('activatePkg', activate_pkg)

  # Save the directory before we change to source root
  exe.env['buildTemp'] = os.getcwd()

  setup.main(exe)

main()
