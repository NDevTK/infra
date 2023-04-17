# Copyright 2016 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Configuration."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import os

from google.appengine.ext import vendor

# Add libraries installed in the lib/ folder.
lib_path = os.path.join(os.path.dirname(os.path.realpath(__file__)), 'lib')
vendor.add(lib_path)
# Add libraries to pkg_resources working set to find the distribution.
import pkg_resources
pkg_resources.working_set.add_entry(lib_path)

import six
reload(six)

import httplib2
import oauth2client

# Only need this for local development. gae_ts_mon.__init__.py inserting
# protobuf_dir to front of sys.path seems to cause this problem.
# See go/monorail-import-mystery for more context.
import settings
if settings.local_mode:
  from google.rpc import status_pb2

from components import utils
utils.import_third_party()
utils.fix_protobuf_package()
