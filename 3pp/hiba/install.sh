#!/bin/bash
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e # Exit immediately if a command exits with a non-zero status.
set -x # Print commands and their arguments as they are executed.
set -o pipefail

PREFIX="$1"
DEPS_PREFIX="$2"

./autogen.sh

export LIBRARY_PATH="${DEPS_PREFIX}/lib"

./configure \
  --disable-shared \
  --with-opensshdir="${DEPS_PREFIX}/include/openssh" \
  --with-openssldir="${DEPS_PREFIX}" \
  --prefix="${PREFIX}" \
  --host="${CROSS_TRIPLE}" \
  CFLAGS="-Wno-unused-command-line-argument" \
  || cat config.log

make install
