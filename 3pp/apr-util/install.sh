#!/bin/bash
# Copyright 2021 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"
DEPS_PREFIX="$2"

./configure --prefix="$PREFIX" --with-apr="$DEPS_PREFIX" --with-expat="$DEPS_PREFIX"
make "-j$(nproc)" LT_LDFLAGS=-static
make install
