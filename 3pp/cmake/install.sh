#!/bin/bash
# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

mkdir cmake-build
cd cmake-build

# Use the cmake in path to bootstrap cmak'ing cmake!
# Force CMAKE_CXX_STANDARD to 14 because C++17 is buggy on gcc10.
# See also: https://github.com/scikit-build/cmake-python-distributions/issues/221
cmake .. \
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
  # Unset CMAKE_TOOLCHAIN_FILE to avoid using host cmake libraries in tests
  env -u CMAKE_TOOLCHAIN_FILE \
    ./bin/ctest --parallel "$(nproc)" \
    --force-new-ctest-process \
    --stop-on-failure \
    --output-on-failure \
    --exclude-regex '(CTestLimitDashJ|BootstrapTest)' # CTestLimitDashJ doesn't work well with parallel
fi

# Use the system cmake to actually do the install. Otherwise it will use the
# cmake binary we just built and fail when we're cross compiling.
cmake .. -P cmake_install.cmake
