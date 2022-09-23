#!/bin/bash
# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
#
# set -e
# set -x
# set -o pipefail
#
PREFIX="$1"

echo $PATH
echo `which cl`
echo `which ml64`

cd CPP/7zip/Bundles/Alone2
nmake PLATFORM=x64

if [ -d x86 ]
then
  cp x86/7zz.exe $PREFIX
else
  cp x64/7zz.exe $PREFIX
fi
