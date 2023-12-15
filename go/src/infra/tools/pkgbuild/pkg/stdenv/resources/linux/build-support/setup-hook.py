# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Setup script with linux specified hooks."""
# pylint: disable=global-at-module-level
# pylint: disable=undefined-variable

def setup(exe) -> None:
  """Build hooks for linux."""
  import os
  import subprocess

  dependencies = []
  native_envs = {}

  def pre_unpack(exe) -> bool:
    envs = subprocess.check_output([
        'docker', 'run', '--rm',
        exe.env['dockerImage'],
        '/usr/bin/env',
    ])
    native_envs.update(e.decode().split('=', 1) for e in envs.splitlines())
    return True

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
      # Exclude dependencies from builtin:import. The path from the host is not
      # valid inside the container.
      stamp = os.path.join(dep, 'build-support', 'base_import.stamp')
      if not os.path.exists(stamp):
        volumes.extend(('--volume', f'{dep}:{dep}'))

    docker = [
        'docker', 'run', '--rm',
        '--workdir', cwd,
        '--user', f'{os.getuid()}:{os.getgid()}',
        '--env', f'HOME={cwd}',
    ]

    env = []
    for k, v in exe.env.items():
      # These environment variables should have the new paths appended in the
      # container.
      if k in {'PATH'} and k in native_envs:
        v = os.path.pathsep.join([v, native_envs[k]])
      env.extend(('--env', f'{k}={v}'))
    # Force override LDFLAGS even it's not set. This is because dockcross by
    # default set it to '-L/usr/cross/lib', which may override the library
    # path passed to the configure.
    if 'LDFLAGS' not in exe.env:
      env.extend(('--env', 'LDFLAGS='))

    impage = [
        exe.env['dockerImage'],
    ]

    subprocess.check_call(
        docker + volumes + env + impage + ['/start.sh'] + ctx.args)
    return True

  def activate_pkg(exe) -> bool:
    ctx = exe.current_context
    dependencies.append(str(ctx.pkg))
    return True

  exe.add_hook('preUnpack', pre_unpack)
  exe.add_hook('executeCmd', execute_cmd)
  exe.add_hook('activatePkg', activate_pkg)

  # Save the directory before we change to source root
  exe.env['buildTemp'] = os.getcwd()

setup(exe)
