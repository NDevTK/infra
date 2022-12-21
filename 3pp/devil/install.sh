#!/bin/bash
# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

PREFIX="$1"

# All of devil along with its dependencies in catapult.
declare -a target_dirs=(
  ".vpython"
  ".vpython3"
  "common/py_utils"
  "dependency_manager"
  "devil"
  "third_party/gsutil"
  "third_party/six"
)

# Use "--parents" to reserve the relative paths, e.g.
# common/py_utils -> $PREFIX/common/py_utils
for target_dir in "${target_dirs[@]}"; do
  cp -rf --parents "$target_dir" "$PREFIX"/
done
