# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Configuration."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import os
import sys

import settings

# Enable third-party imports
from google.appengine.ext import vendor
vendor.add(os.path.join(os.path.dirname(__file__), 'third_party'))

# Set path to your libraries folder.
lib_path = os.path.join(os.path.dirname(os.path.realpath(__file__)), 'lib')

# Add libraries installed in the path folder.
vendor.add(lib_path)


# Add libraries to pkg_resources working set to find the distribution.
import pkg_resources
pkg_resources.working_set.add_entry(lib_path)

import six
reload(six)

import httplib2
import oauth2client

from components import utils
utils.fix_protobuf_package()

# Only import in non-test modes because testing environments do not have grpc,
# a binary package used by cloud tasks.
# Only import after fixing protobuf package because appengine python 2 only
# includes protobuf 3.0.0 alpha, but cloud tasks via googleapis-common-protos
# requires minnimum of protobuf 3.6.0. fix_protobuf_package() inserts protobuf
# version 3.12.1
# https://source.chromium.org/chromium/_/chromium/infra/luci/luci-py/+/HEAD:client/third_party/google/protobuf/__init__.py;l=33

if not settings.unit_test_mode:
  from google import protobuf
  logging.info(protobuf.__version__)
  from google.cloud import tasks
