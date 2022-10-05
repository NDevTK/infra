#!/bin/bash
# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

set -e
set -x
set -o pipefail

# CMake's tests may be affected by the environment variables. Using `env it` to
# isolate from dockcross environments.
env -i PATH="$PATH" ./bin/ctest -j "$(nproc)" \
  --force-new-ctest-process \
  --stop-on-failure \
  --output-on-failure \
  -E BootstrapTest