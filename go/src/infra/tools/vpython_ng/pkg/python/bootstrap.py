# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import glob
import os
import subprocess
import sys

# Create virtual environment in ${out} directory
virtualenv = glob.glob(
    os.path.join(os.environ['virtualenv'], '*', 'virtualenv.py'))[0]
subprocess.check_call([
    sys.executable, virtualenv, '--no-download', '--always-copy',
    os.environ['out']
])

# Install wheels to virtual environment
if 'wheels' in os.environ:
  pip = glob.glob(os.path.join(os.environ['out'], '*', 'pip3*'))[0]
  subprocess.check_call([
      pip,
      'install',
      '--isolated',
      '--compile',
      '--no-index',
      '--find-links',
      os.path.join(os.environ['wheels'], 'wheels'),
      '--requirement',
      os.path.join(os.environ['wheels'], 'requirements.txt'),
  ])

# Generate all .pyc in the output directory. This prevent generating .pyc on the
# fly, which modifies derivation output after the build.
subprocess.check_call([
    sys.executable, '-m', 'compileall', os.environ['out']
])
