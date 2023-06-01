#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

if [[ $_3PP_PLATFORM == $_3PP_TOOL_PLATFORM ]]; then # not cross-compiling
  nmake -f makefile.msc test
fi
nmake -f makefile.msc lib bzip2

mkdir "${PREFIX}/bin"
cp bzip2.exe "${PREFIX}/bin"
# Install the executable to other names as well, for compatibility
# with Linux/Mac.
cp bzip2.exe "${PREFIX}/bin/bunzip2.exe"
cp bzip2.exe "${PREFIX}/bin/bzcat.exe"

mkdir "${PREFIX}/include"
cp bzlib.h "${PREFIX}/include"

mkdir "${PREFIX}/lib"
cp *.lib "${PREFIX}/lib"
