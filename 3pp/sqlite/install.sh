#!/bin/bash
# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"
export CFLAGS="${CFLAGS} -fPIC"

./configure \
  --prefix "$PREFIX" \
  --host "$CROSS_TRIPLE" \
  --enable-static \
  --disable-shared \
  --with-pic \
  --enable-fts5 \
  --enable-json1 \
  --enable-session

make install -j $(nproc)
