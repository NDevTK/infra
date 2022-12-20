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
  "common/py_utils"
  "dependency_manager"
  "devil"
  "third_party/gsutil"
)

for target_dir in "${target_dirs[@]}"; do
  cp -rf  "$target_dir" "$PREFIX"/
done
