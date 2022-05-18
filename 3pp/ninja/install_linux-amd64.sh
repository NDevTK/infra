#!/bin/bash
# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"
BASEDIR=$(pwd)

# Fetch Chromium's sysroot
CHROMIUM_BUILD_REVISION="28bea73326715ae8bc8673b16046d0c32df48a3e"
mkdir chromium_build
(
  cd chromium_build &&
  git init . &&
  git remote add origin https://chromium.googlesource.com/chromium/src/build &&
  git fetch --depth=1 origin "$CHROMIUM_BUILD_REVISION" &&
  git checkout "$CHROMIUM_BUILD_REVISION" &&
  ./linux/sysroot_scripts/install-sysroot.py --arch=x64
)
SYSROOT="${BASEDIR}/chromium_build/linux/debian_bullseye_amd64-sysroot"
LIB_PATH="${SYSROOT}/usr/lib/x86_64-linux-gnu"
INCLUDE_PATH="${SYSROOT}/usr/include/x86_64-linux-gnu"

# Fetch Chromium's clang
CHROMIUM_CLANG_REVISION="6e492e7a5c4b7c7d8a59a5568d57d436e17c28e9"
curl "https://chromium.googlesource.com/chromium/src/tools/clang/+/${CHROMIUM_CLANG_REVISION}/scripts/update.py?format=TEXT" \
  | base64 -d | python3 - --output-dir=chromium_clang
CLANG_PATH="${BASEDIR}/chromium_clang/bin/clang++"

CXX="$CLANG_PATH" \
CFLAGS="--sysroot=${SYSROOT} -I${INCLUDE_PATH}" \
LDFLAGS="--sysroot=${SYSROOT} -B${LIB_PATH}" \
  ./configure.py --bootstrap
./ninja all
./ninja_test
cp ninja "$PREFIX"
