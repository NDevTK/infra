#!/bin/bash
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e # Exit immediately if a command exits with a non-zero status.
set -x # Print commands and their arguments as they are executed.
set -o pipefail

PREFIX="$1"
DEPS_PREFIX="$2"

autoreconf

./configure \
  --with-ssl-dir="${DEPS_PREFIX}" \
  --prefix="${PREFIX}" \
  --host="${CROSS_TRIPLE}" \
  || cat config.log

make

mkdir -p "${PREFIX}/lib"
cp libssh.a "${PREFIX}/lib/"

mkdir -p "${PREFIX}/include/openssh"
cp -r *.h "${PREFIX}/include/openssh/"

