#!/bin/bash
# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

# Running tests in 
env -i PATH="$PATH" ./bin/ctest -j "$(nproc)" \
  --force-new-ctest-process \
  --stop-on-failure \
  --output-on-failure \
  -E BootstrapTest