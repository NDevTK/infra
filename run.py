#!/usr/bin/env vpython3
# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Wrapper for `vpython3 -m` to make running tools simpler.

A tool is defined as a python module with a __main__.py file. This latter file
is run by the present script.

In particular, allows gclient to change directories when running hooks for
infra.
"""

assert __name__ == '__main__'

import importlib.util
import os
import sys

RUNPY_PATH = os.path.abspath(__file__)
ROOT_PATH = os.path.dirname(RUNPY_PATH)

# Do not want to mess with sys.path, load the module directly.
spec = importlib.util.spec_from_file_location(
    'run_helper', os.path.join(ROOT_PATH, 'bootstrap', 'run_helper.py'))
run_helper = importlib.util.module_from_spec(spec)
spec.loader.exec_module(run_helper)

sys.exit(run_helper.run_py_main(sys.argv[1:], RUNPY_PATH, 'infra'))
