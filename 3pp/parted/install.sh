#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e # Exit immediately if a command exits with a non-zero status.
set -x # Print commands and their arguments as they are executed.
set -o pipefail

PREFIX="$1"
DEPS="$2"
# Lib path for libuuid
export LDFLAGS="$LDFLAGS -L$DEPS/lib"
# libuuid headers
export CFLAGS="$CFLAGS -I$DEPS/include"
# fix for pkg.m4 (probably a bug as it should check the default locations)
export ACLOCAL_FLAGS="$ACLOCAL_FLAGS -I/usr/share/aclocal"
./bootstrap
# disable device mapper as we don't plan on using this on physical hardware.
# disable readline as we don't need it and it requires extra dependency.
./configure --prefix=${PREFIX} \
  --disable-device-mapper \
  --without-readline \
  --disable-shared \
  --enable-static

make install -j$(nproc)
