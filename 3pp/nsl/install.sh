#!/bin/bash
# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

# We don't want to link against 'libtirpc' in the Docker container
# because that will cause it to become a dependency for libnsl.
# There is no clean way to disable this configure check, so instead
# we intentionally break pkg-config.
export PKG_CONFIG_LIBDIR=/invalid

./configure --disable-shared \
  --prefix "$PREFIX" \
  --host "$CROSS_TRIPLE"
make install -j $(nproc)
