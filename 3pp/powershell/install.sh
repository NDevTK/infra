#!/bin/bash
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

if [[ $_3PP_PLATFORM = "windows-arm64" ]]; then
  cp *.zip $PREFIX/powershell.zip
else
  cp *.msi $PREFIX/powershell.msi
fi
