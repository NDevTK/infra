#!/bin/bash
# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

# Use the cmake in path to bootstrap cmak'ing cmake!
cmake \
  -GNinja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$PREFIX" \
  -DCMAKE_USE_OPENSSL=OFF \
  -DBUILD_TESTING:BOOL=ON \
  -DCMAKE_CXX_STANDARD:STRING=14
# Our dockcross environment should automatically set the CMAKE toolchain to
# enable cross-compilation.

# Build all the stuffs.
cmake --build . -j "$(nproc)"

# Run the test suite, if not cross-compiling.
if [[ $_3PP_PLATFORM == "$_3PP_TOOL_PLATFORM" ]]; then
  # CMake's tests may be affected by the environment variables. Using `env it`
  # to isolate tests from dockcross environments.
  env -i PATH="$PATH" ./bin/ctest -j "$(nproc)" \
    --force-new-ctest-process \
    --stop-on-failure \
    --output-on-failure \
    -E BootstrapTest
fi

# Use the system cmake to actually do the install. Otherwise it will use the
# cmake binary we just built and fail when we're cross compiling.
cmake -P cmake_install.cmake
