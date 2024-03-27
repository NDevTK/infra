#!/bin/bash
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

nmake -f win32/Makefile.msc zlib.lib

mkdir -p "${PREFIX}/include"
cp zconf.h "${PREFIX}/include/"
cp zlib.h "${PREFIX}/include/"

mkdir -p "${PREFIX}/lib"
cp zlib.lib  "${PREFIX}/lib/"
