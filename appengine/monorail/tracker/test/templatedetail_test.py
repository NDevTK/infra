# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Unit tests for Template editing/viewing servlet."""


import unittest

import settings

from framework import permissions
from services import service_manager
from testing import fake
from testing import testing_helpers
from tracker import templatedetail


class TemplateDetailTest(unittest.TestCase):
  """Tests for the TemplateDetail servlet."""

  def setUp(self):
    self.cnxn = 'fake cnxn'
    self.mr = testing_helpers.MakeMonorailRequest()
    self.services = service_manager.Services(project=fake.ProjectService(),
                                             config=fake.ConfigService(),
                                             user=fake.UserService())
    self.servlet = templatedetail.TemplateDetail('req', 'res',
                                               services=self.services)

  def testAssertBasePermission(self):
    pass

  def testGatherPageData(self):
    pass

  def testProcessFormData(self):
    pass
