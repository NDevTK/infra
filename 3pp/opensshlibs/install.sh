#!/bin/bash
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e # Exit immediately if a command exits with a non-zero status.
set -x # Print commands and their arguments as they are executed.
set -o pipefail

PREFIX="$1"
DEPS_PREFIX="$2"

./configure \
  --with-ssl-dir="${DEPS_PREFIX}" \
  --prefix="${PREFIX}" \
  --host="${CROSS_TRIPLE}" \
  || cat config.log

make libssh.a openbsd-compat/libopenbsd-compat.a

# OpenSSH does not export *.a and *.h files with a make install. So we have to
# copy them instead.

mkdir -p "${PREFIX}/lib"
cp libssh.a "${PREFIX}/lib/"
cp openbsd-compat/libopenbsd-compat.a "${PREFIX}/lib/"

mkdir -p "${PREFIX}/include/openssh/openbsd-compat"
cp -r *.h "${PREFIX}/include/openssh/"
cp -r openbsd-compat/*.h "${PREFIX}/include/openssh/openbsd-compat/"
