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
  -DCMAKE_BUILD_TYPE:STRING=Release \
  -DCMAKE_INSTALL_PREFIX:STRING="$PREFIX" \
  -DCMAKE_USE_OPENSSL:BOOL=OFF \
  -DBUILD_TESTING:BOOL=ON \
  -DCMAKE_CXX_STANDARD:STRING=14
# Our dockcross environment should automatically set the CMAKE toolchain to
# enable cross-compilation.

# Build all the stuffs.
cmake --build . -j "$(nproc)"

# Run the test suite, if not cross-compiling.
# TODO(fancl): Fix tests on windows.
if [[ "$_3PP_PLATFORM" == "$_3PP_TOOL_PLATFORM" && "$_3PP_PLATFORM" != windows-* ]]; then
  # RunCMake.CPack_STGZ (Self extracting Tar GZip compression) will generate a
  # shell script use pax instead of tar if possible to unpack itself. This is
  # fine in most cases since STGZ use restricted-pax format which is supported
  # by pax, but may break if libarchive considers restricted-pax can't store the
  # metadata (e.g. UID/GID too big) and uses unrestricted pax format instead.
  # Unfortunately our UID do exceed the limit, which will result in unpacking
  # extra PaxHeader in the unittest.
  #
  # CTestLimitDashJ doesn't work well with parallel.
  # FileDownload can be flaky in parallel because it relies on execution order.
  #
  # Unset CMAKE_TOOLCHAIN_FILE to avoid using host cmake libraries in tests
  env -u CMAKE_TOOLCHAIN_FILE \
    ./bin/ctest --parallel "$(nproc)" \
    --force-new-ctest-process \
    --stop-on-failure \
    --output-on-failure \
    --exclude-regex '(RunCMake.CPack_STGZ|CTestLimitDashJ|FileDownload|BootstrapTest)'

  env -u CMAKE_TOOLCHAIN_FILE \
    ./bin/ctest \
    --force-new-ctest-process \
    --stop-on-failure \
    --output-on-failure \
    --tests-regex '(CTestLimitDashJ|FileDownload)'
fi

# Use the system cmake to actually do the install. Otherwise it will use the
# cmake binary we just built and fail when we're cross compiling.
cmake .. -P cmake_install.cmake
