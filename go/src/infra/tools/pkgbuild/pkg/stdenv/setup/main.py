# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Helper script for updating search path."""
# pylint: disable=global-at-module-level
# pylint: disable=undefined-variable

import os
import sys


def _stdenv_path() -> str:
  file_path = os.path.realpath(__file__)  # {stdenv}/setup/main.py
  return os.path.dirname(os.path.dirname(file_path))


# Insert the path to stdenv for python searching setup module.
sys.path.insert(0, _stdenv_path())

import setup

setup.main()
