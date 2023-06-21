# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

# pylint: disable=undefined-variable

"""Setup expect_tests environment.

Derived from Monorail's version of this file. See also
https://chromium.googlesource.com/infra/testing/expect_tests/
"""

import os
import sys

# Using pretest_filename is magic, because it is available in the locals() of
# the script which execfiles this file.
# prefixing with 'pretest' to avoid name collisions in expect_tests.
pretest_APPENGINE_ENV_PATH = os.path.join(
    os.path.dirname(
        os.path.dirname(
            os.path.dirname(
                  os.path.abspath(pretest_filename)))),
    'cipd', 'gcloud', 'platform', 'google_appengine')
sys.path.append(pretest_APPENGINE_ENV_PATH)

# Crazy hack, because of appengine.
# Importing dev_appserver is probably not officially supported and fix_sys_path
# may be an implementation detail subject to change.
import dev_appserver as pretest_dev_appserver

pretest_dev_appserver.fix_sys_path()

# Remove google_appengine SDK from sys.path after use
sys.path.remove(pretest_APPENGINE_ENV_PATH)

SDK_LIBRARY_PATHS = [
  # This is not added by fix_sys_path.
  os.path.join(pretest_APPENGINE_ENV_PATH, 'lib', 'mox'),
]
sys.path.extend(SDK_LIBRARY_PATHS)

# Add pylib to the path
sys.path.append(os.path.join(os.path.dirname(os.path.abspath(pretest_filename)),
                             'pylib'))

os.environ['SERVER_SOFTWARE'] = 'test ' + os.environ.get('SERVER_SOFTWARE', '')
os.environ['CURRENT_VERSION_ID'] = 'test.123'
os.environ.setdefault('NO_GCE_CHECK', 'True')
