# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Setup script for adapting 3pp spec."""
# pylint: disable=global-at-module-level
# pylint: disable=undefined-variable


# Phases


global configure_phase
global build_phase
global install_phase


def configure_phase(_) -> None:
  return


def build_phase(exe) -> None:
  """Run build command in the source directory."""
  import json
  import os

  args = json.loads(exe.env['fromSpecInstall'])
  args.append(exe.env['out'])
  args.append(exe.env['_3PP_PREFIX'])
  _, ext = os.path.splitext(args[0])
  args.insert(0, 'python3' if ext == 'py' else 'bash')

  exe.execute_cmd(args)


def install_phase(_) -> None:
  return


# Hooks


def setup(exe):
  """Copy all libraries into a single directory."""
  import itertools
  import os
  import shutil

  d = os.path.join(os.getcwd(), '3pp_prefix')
  os.makedirs(d)
  exe.env['_3PP_PREFIX'] = d

  def pre_unpack(exe) -> bool:

    def pkgs(name: str) -> List[str]:
      if e := exe.env.get(name):
        return e.split(os.path.pathsep)
      return []

    for pkg in itertools.chain(pkgs('depsHostHost'), pkgs('depsHostTarget')):
      shutil.copytree(pkg, exe.env['_3PP_PREFIX'],
                      symlinks=True, dirs_exist_ok=True)
    return True

  exe.add_hook('preUnpack', pre_unpack)

setup(exe)
