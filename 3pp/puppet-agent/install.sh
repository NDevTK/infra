#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

if [[ $_3PP_PLATFORM == windows* ]]; then
  cp *.msi $PREFIX/puppet-agent-7.msi
fi
