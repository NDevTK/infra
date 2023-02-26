#!/bin/bash
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PREFIX="$1"

nmake -f Mkfiles/msvc.mak
mkdir -p "${PREFIX}"/bin
cp nasm.exe ndisasm.exe "${PREFIX}"/bin

