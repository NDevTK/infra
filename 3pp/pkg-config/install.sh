#!/bin/bash
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e # Exit immediately if a command exits with a non-zero status.
set -x # Print commands and their arguments as they are executed.
set -o pipefail

PREFIX="$1"
DEPS_PREFIX="$2"

_CONFIG_ARGS=(
  "CFLAGS=-O2"
  "--disable-shared"
  "--prefix=${PREFIX}"
)

if [[ -n "${CROSS_TRIPLE}" ]]; then
  _CONFIG_ARGS=( "${_CONFIG_ARGS[@]}" "--host=${CROSS_TRIPLE}" )
fi

echo $PATH

ls -l


./autogen.sh
./configure "${_CONFIG_ARGS[@]}" || cat config.log
make -j $(nproc) install
