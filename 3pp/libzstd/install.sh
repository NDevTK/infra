#!/bin/bash
# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

# clock_gettime() moved to libc in glibc 2.17, however our linux-armv6
# image is on a much older glibc 2.13 that has it in -lrt.
if [[ $_3PP_PLATFORM == linux-armv6l ]]; then
  export LDLIBS=-lrt
fi

make -j $(nproc) install prefix="$PREFIX"
