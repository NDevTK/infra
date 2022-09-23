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

cd CPP/7zip/Bundles/Alone2
make -j -f ../../cmpl_gcc.mak
cp b/c/7zz $PREFIX
