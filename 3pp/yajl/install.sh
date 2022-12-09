#!/bin/bash
# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

# Run: https://github.com/lloyd/yajl/blob/master/configure
./configure -p $PREFIX
make -j $(nproc) install
