# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Default setup script."""
# pylint: disable=global-at-module-level
# pylint: disable=undefined-variable

import sys

# Insert the path to stdenv for python searching setup module.
sys.path.insert(0, sys.argv[1])

import setup

setup.main()
