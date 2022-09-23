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
CHECKOUT_DIR=$(pwd)

cd $CHECKOUT_DIR/CPP/7zip/Bundles/Alone
make -j -f ../../cmpl_clang.mak
cp b/g/7za $PREFIX

cd $CHECKOUT_DIR/CPP/7zip/Bundles/Alone2
make -j -f ../../cmpl_clang.mak
cp b/g/7zz $PREFIX

cd $CHECKOUT_DIR/CPP/7zip/Bundles/Alone7z
make -j -f ../../cmpl_clang.mak
cp b/g/7zr $PREFIX
