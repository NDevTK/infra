#!/bin/bash
# Copyright 2021 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

export CXXFLAGS+=" -fPIC"

PROTOC_OPT=
if [[ $_3PP_PLATFORM != $_3PP_TOOL_PLATFORM ]]; then  # cross compiling
  BUILD_PROTOC=`which protoc`
  PROTOC_OPT=-DWITH_PROTOC=${BUILD_PROTOC}
fi

mkdir cmake-build
cd cmake-build
cmake .. \
  -DCMAKE_BUILD_TYPE:STRING=Release \
  -DCMAKE_INSTALL_PREFIX:STRING="${PREFIX}" \
  -DCMAKE_CXX_STANDARD=14 \
  ${PROTOC_OPT}

make -j $(nproc)
if [[ $_3PP_PLATFORM == $_3PP_TOOL_PLATFORM ]]; then
  make test -j $(nproc)
fi
make install -j $(nproc)
