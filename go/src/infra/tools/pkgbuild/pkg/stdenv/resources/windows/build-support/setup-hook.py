# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Setup script with linux specified hooks."""
# pylint: disable=global-at-module-level
# pylint: disable=undefined-variable

import sys
import subprocess


def setup(exe) -> None:
  """Build hooks for windows."""

  def execute_cmd(exe) -> bool:
    ctx = exe.current_context
    args = ctx.args.copy()

    try:
      subprocess.check_call(args, env=exe.env)
      return True
    except OSError as e:
      print(f'failed to call: {args[0]}: {e}', flush=True)

    try:
      args[0] = args[0] + '.lnk'
      subprocess.check_call(ctx.args, env=exe.env, shell=True)
      return True
    except OSError as e:
      print(f'failed to call: {args[0]}: {e}', flush=True)

    raise OSError('failed to execute cmd')

  exe.add_hook('executeCmd', execute_cmd)


setup(exe)
