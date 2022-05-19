#!/bin/bash
# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"
BASEDIR=$(pwd)

# Fetch Chromium's clang
CHROMIUM_CLANG_REVISION="6e492e7a5c4b7c7d8a59a5568d57d436e17c28e9"
curl "https://chromium.googlesource.com/chromium/src/tools/clang/+/${CHROMIUM_CLANG_REVISION}/scripts/update.py?format=TEXT" \
  | base64 -d | python3 - --output-dir=chromium_clang
CLANG_PATH="${BASEDIR}/chromium_clang/bin/clang++"

# Find Xcode SDK
XCODE_SDK_PATH=$(xcrun --show-sdk-path)
MACOSX_VERSION_MIN="10.13"

CXX="$CLANG_PATH" \
CFLAGS="-isysroot${XCODE_SDK_PATH} -mmacosx-version-min=${MACOSX_VERSION_MIN}" \
LDFLAGS="${CFLAGS}" \
  ./configure.py --bootstrap
./ninja all
./ninja_test
cp ninja "$PREFIX"
