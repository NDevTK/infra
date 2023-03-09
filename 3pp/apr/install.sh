#!/bin/bash
# Copyright 2021 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

CFLAGS="-g -O2 -Wno-format -Wno-implicit-function-declaration"

CFLAGS="$CFLAGS" ./configure --prefix="$PREFIX"
make "-j$(nproc)"
make install
# Remove config files that have hard-coded paths.
rm -rf $PREFIX/lib/pkgconfig
rm -f $PREFIX/lib/*.la
# Replace hard-coded sed path.
sed -i 's/^SED=.*/SED="sed"/' "$PREFIX/build-1/libtool"
