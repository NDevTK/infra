#!/bin/bash
# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

# TODO(fancl): Add support for host dependencies in 3pp spec.
# We need to explicitly import host sed because we won't rely on PATH in
# the new system.
function sed {
  if [[ -f /usr/bin/sed ]]; then
    /usr/bin/sed "$@"
  else
    /bin/sed "$@"
  fi
}
export -f sed

./configure --enable-static --disable-shared \
  --prefix "$PREFIX" \
  --host "$CROSS_TRIPLE"
make install -j $(nproc)
