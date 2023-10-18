# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Setup script for adapting 3pp spec."""
# pylint: disable=global-at-module-level
# pylint: disable=undefined-variable


#############################################################################
# Phases
#############################################################################

global configure_phase
global build_phase
global install_phase


def configure_phase(_) -> None:
  return


def build_phase(exe) -> None:
  """Run build command in the source directory."""
  import json
  import os
  import shutil

  # Copy source to output if no install section presented in the spec.
  if exe.env.get('_3PP_NO_INSTALL'):
    shutil.copytree(os.getcwd(), exe.env['out'],
                    symlinks=True, dirs_exist_ok=True)
    return

  args = json.loads(exe.env['fromSpecInstall'])
  args[0] = os.path.join(exe.env['_3PP_DEF'], args[0])
  args.append(exe.env['out'])
  args.append(exe.env['_3PP_PREFIX'])
  _, ext = os.path.splitext(args[0])
  args.insert(0, 'python3' if ext == 'py' else 'bash')

  exe.execute_cmd(args)


def install_phase(_) -> None:
  return

#############################################################################
# Hooks
#############################################################################

def setup(exe):
  """Copy all libraries into a single directory."""
  import itertools
  import os
  import shutil

  d = os.path.join(os.getcwd(), '3pp_prefix')
  exe.env['_3PP_PREFIX'] = d

  # Workaround for packages putting executable files under the root directory.
  # e.g. ninja, pip_bootstrap
  def activate_pkg(exe) -> bool:
    ctx = exe.current_context
    if ctx.host == PlatType.BUILD and not ctx.pkg.joinpath('bin').is_dir():
      exe.append_to_search_path(Execution.ENV_PATH, ctx.pkg)
    return True

  # Copy the source before unpack if _3PP_UNPACK_ARCHIVE is not set.
  def skip_unpack(exe):
    if exe.env.get('_3PP_UNPACK_ARCHIVE'):
      return False
    return extract.copy_cmd(exe)

  # We don't actually flatten the directories. We chdir into the single
  # directories instead to Simulate the behaviour of
  # api.file.flatten_single_directories recipe.
  def set_source_root(exe) -> bool:
    root = os.getcwd()
    if exe.env.get('_3PP_NO_ARCHIVE_PRUNE'):
      exe.env[exe.ENV_SOURCE_ROOT] = root
      return True

    # Move the root if it only contains a single directory.
    while len(names := os.listdir(root)) == 1 and os.path.isdir(names[0]):
      root = names[0]

    exe.env[exe.ENV_SOURCE_ROOT] = root
    return True

  def post_unpack(exe) -> bool:
    def pkgs(name: str) -> List[str]:
      if e := exe.env.get(name):
        return e.split(os.path.pathsep)
      return []

    for pkg in itertools.chain(pkgs('depsHostHost'), pkgs('depsHostTarget')):
      shutil.copytree(pkg, exe.env['_3PP_PREFIX'],
                      symlinks=True, dirs_exist_ok=True)
    return True

  exe.add_hook('activatePkg', activate_pkg)
  exe.add_hook('unpackCmd', skip_unpack)
  exe.add_hook('setSourceRoot', set_source_root)
  exe.add_hook('postUnpack', post_unpack)

setup(exe)
