#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e # Exit immediately if a command exits with a non-zero status.
set -x # Print commands and their arguments as they are executed.
set -o pipefail

PREFIX="$1"
# disable features that need external dependencies. We only need rsync for
# compiling parted
./configure --prefix=${PREFIX} \
  --disable-md2man \
  --disable-openssl \
  --disable-xxhash \
  --disable-zstd \
  --disable-lz4
make install -j$(nproc)
