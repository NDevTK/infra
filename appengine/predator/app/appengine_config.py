# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
import pkg_resources

import site
import six
import sys


_THIS_DIR = os.path.realpath(os.path.dirname(__file__))
_LIB = os.path.join(_THIS_DIR, 'lib')

site.addsitedir(_LIB)
sys.path.append(_LIB)
pkg_resources.working_set.add_entry(_LIB)

# Add all the first-party and third-party libraries.
sys.path.append(os.path.realpath(os.path.join(_THIS_DIR, 'first_party')))
sys.path.append(os.path.realpath(os.path.join(_THIS_DIR, 'third_party')))
sys.path.append(
    os.path.realpath(
        os.path.join(_THIS_DIR, 'third_party', 'pipeline', 'python', 'src')))

if six.PY2:
  sys.path.append(
      os.path.realpath(os.path.join(_THIS_DIR, 'third_party', 'python2')))
