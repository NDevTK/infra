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
  "--enable-static"
  "--prefix=${PREFIX}"
)

if [[ -n "${CROSS_TRIPLE}" ]]; then
  _CONFIG_ARGS=( "${_CONFIG_ARGS[@]}" "--host=${CROSS_TRIPLE}" )
fi

echo $PATH

ls -l

# autogen.sh does `LIBTOOLIZE=${LIBTOOLIZE:-glibtoolize}` when uname reports
# Darwin, but libtool/libtoolize brought in as a dependency in 3pp.pb is already
# the GNU one. For Linux, this is effectively a no-op since autogen.sh sets it
# to libtoolize when any other OS is detected.
# export LIBTOOLIZE=libtoolize
# ./autogen.sh

mkdir build
pushd build

../configure --help


../configure "${_CONFIG_ARGS[@]}" || cat config.log
make -j $(nproc) install
