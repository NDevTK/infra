#!/bin/bash
# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

./configure --enable-static --disable-shared \
  --prefix "$PREFIX" \
  --host "$CROSS_TRIPLE"
make install -j $(nproc)

mkdir -p "$PREFIX/build-support"
cat > "$PREFIX/build-support/setup-hook.py" << EOF
def setup(exe):
  import os

  def activate_pkg(exe) -> bool:
    ctx = exe.current_context
    if (aclocal := ctx.pkg.joinpath('share', 'aclocal')).is_dir():
      exe.append_to_search_path('ACLOCAL_PATH', aclocal)
    return True

  exe.add_hook('activatePkg', activate_pkg)
  pass

setup(exe)
EOF
