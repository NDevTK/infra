#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

cp -rf --parents * "$PREFIX"/

# Copy swig.exe to bin/ so that it can be found in the same location
# on all platforms.
mkdir "$PREFIX"/bin
cp swig.exe "$PREFIX"/bin
