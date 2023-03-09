#!/bin/bash
# Copyright 2021 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"
export CFLAGS="-fPIC"

./configure --enable-static=yes --enable-shared=no \
  --with-device=ch3 --disable-fortran \
  --prefix="$PREFIX" \
  --host="$CROSS_TRIPLE"
make install -j $(nproc)
