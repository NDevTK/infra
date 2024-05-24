#!/bin/bash
# Copyright 2021 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

export CXXFLAGS+=" -fPIC"

# On RISC-V, some runtime functions are in libatomic instead of libc.
if [[ $_3PP_PLATFORM == "linux-riscv64" ]]; then
  export LDFLAGS+=" -latomic"
fi

PROTOC_OPT=
if [[ $_3PP_PLATFORM != $_3PP_TOOL_PLATFORM ]]; then  # cross compiling
  PROTOC_OPT="-Dprotobuf_BUILD_TESTS=OFF"
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
