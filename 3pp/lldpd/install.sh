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

case "${_3PP_PLATFORM}" in
  mac-*)
    _CONFIG_ARGS=(
      "${_CONFIG_ARGS[@]}"
      "--with-launchddaemonsdir=${PREFIX}/Library/LaunchDaemons"
    )
    ;;
  linux-*)
    _CONFIG_ARGS=(
      "${_CONFIG_ARGS[@]}"
      "--with-systemdsystemunitdir=${PREFIX}/usr/lib/systemd/system"
      "--with-sysusersdir=${PREFIX}/usr/lib/sysusers.d"
    )
    ;;
esac


echo $PATH

ls -l

mkdir build
pushd build

../configure --help

../configure "${_CONFIG_ARGS[@]}" || cat config.log

ls -l

cat Makefile

make -j $(nproc) install

ls -lR ${PREFIX}
