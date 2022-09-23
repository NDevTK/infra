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
PWD="$(PWD)"

# Move /usr/bin to the end of PATH because otherwise nmake will use
# /usr/bin/link, which doesn't work, instead of the MSVC linker.
PATH=$(echo $PATH | sed 's/:\/usr\/bin//g'):/usr/bin

cd $PWD/CPP/7zip/Bundles/Alone
nmake PLATFORM=x86
cp x86/7za.exe $PREFIX

cd $PWD/CPP/7zip/Bundles/Alone2
nmake PLATFORM=x86
cp x86/7zz.exe $PREFIX

cd $PWD/CPP/7zip/Bundles/Alone7z
nmake PLATFORM=x86
cp x86/7zr.exe $PREFIX
