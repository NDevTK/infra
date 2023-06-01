#!/bin/bash
# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

if [[ $_3PP_PLATFORM == $_3PP_TOOL_PLATFORM ]]; then # not cross-compiling
  make test
fi
make install PREFIX="$PREFIX"

