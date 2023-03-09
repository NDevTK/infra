#!/bin/bash
# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"
DEPS="$2"

# Ensure that the static library can be linked into a wheel shared library.
export CFLAGS="${CFLAGS} -fPIC"

# The presence of LDFLAGS in the environment suppresses detection of the
# -pthread flag by CMake, so set it explicitly on Linux.
if [[ $_3PP_PLATFORM == linux* ]]; then
  export LDFLAGS="${LDFLAGS} -pthread"
fi

mkdir build
cmake \
  -Bbuild \
  -DWITH_UNIT_TESTS=OFF \
  -DWITHOUT_SERVER=ON \
  -DWITH_BOOST=boost \
  -DWITH_SSL=${DEPS} \
  -DCMAKE_INSTALL_PREFIX=${PREFIX}

cd build
make -j $(nproc) install
