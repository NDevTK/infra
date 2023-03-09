#!/bin/bash
# Copyright 2019 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"
DEPS_PREFIX=$2

cd ./expat

PATH=$DEPS_PREFIX/bin:$PATH ./buildconf.sh
./configure --prefix="$PREFIX" --enable-shared=no --host "$CROSS_TRIPLE"
make install
# pkg-config will have the original build prefix, which is not useful
# for relocatable packages. Remove the configs completely.
rm -rf $PREFIX/lib/pkgconfig
# libtool library similarly has a hardcoded path.
rm -f $PREFIX/lib/libexpat.la
