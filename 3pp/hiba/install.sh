#!/bin/bash
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e # Exit immediately if a command exits with a non-zero status.
set -x # Print commands and their arguments as they are executed.
set -o pipefail

PREFIX="$1"
DEPS_PREFIX="$2"

ls -l "${DEPS_PREFIX}"
ls -l "${DEPS_PREFIX}/lib"
ls -l "${DEPS_PREFIX}/include"

./autogen.sh

LD_LIBRARY_PATH="${DEPS_PREFIX}/lib" ./configure \
  --with-opensshdir="${DEPS_PREFIX}" \
  --with-openssldir="${DEPS_PREFIX}/include" \
  --disable-shared \
  --prefix="${PREFIX}" \
  --host="${CROSS_TRIPLE}" \
  || cat config.log

make install
