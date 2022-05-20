#!/bin/bash
# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

CFLAGS=""
if [[ $_3PP_PLATFORM == mac* ]]; then
  XCODE_SDK_PATH=$(xcrun --show-sdk-path)
  # the min version is configured in mac_sdk.gni
  # https://crrev.com/e840c4b48a861be294f206bd694ebce986ddbb88/build/config/mac/mac_sdk.gni#28
  MACOSX_VERSION_MIN="10.13"
  CFLAGS="-isysroot${XCODE_SDK_PATH} -mmacosx-version-min=${MACOSX_VERSION_MIN}"
fi

CFLAGS="${CFLAGS}" ./configure.py --bootstrap
./ninja all
./ninja_test
if [[ $_3PP_PLATFORM == windows* ]]; then
  cp ninja.exe "$PREFIX"
else
  cp ninja "$PREFIX"
fi
