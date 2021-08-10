# Copyright (c) 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import pkg_resources
from google.appengine.ext import vendor

# Make webapp.template use django 1.2.  Quells default django warning.
webapp_django_version = '1.2'

path = 'lib'
# Add libraries installed in the path folder.
vendor.add(path)
# Add libraries to pkg_resources working set to find the distribution.
pkg_resources.working_set.add_entry(path)
