# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Setup script with darwin specified hooks."""
# pylint: disable=global-at-module-level
# pylint: disable=undefined-variable

def setup(exe) -> None:
  """Build hooks for darwin."""
  import pathlib

  def pre_unpack(exe) -> bool:
    base = pathlib.Path(exe.env['DEVELOPER_DIR'])
    assert base.exists(), f'xcode base: {base} should exists'
    exe.append_to_search_path(Execution.ENV_PATH, base.joinpath('usr', 'bin'))

    toolchain = base.joinpath('Toolchains', 'XcodeDefault.xctoolchain')
    assert toolchain.exists(), f'xcode toolchain: {toolchain} should exists'
    exe.append_to_search_path(
        Execution.ENV_PATH, toolchain.joinpath('usr', 'bin'))

    sdk = base.joinpath('Platforms', 'MacOSX.platform', 'Developer', 'SDKs',
                        'MacOSX.sdk')
    assert toolchain.exists(), f'xcode sdk: {sdk} should exists'
    exe.append_to_search_path(Execution.ENV_PATH, sdk.joinpath('usr', 'bin'))

    return True

  exe.add_hook('preUnpack', pre_unpack)

setup(exe)
