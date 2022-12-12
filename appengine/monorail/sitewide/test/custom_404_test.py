# Copyright 2017 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Unit tests for the custom_404 servlet."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

from six.moves import http_client
import unittest

from framework import exceptions
from services import service_manager
from sitewide import custom_404
from testing import fake
from testing import testing_helpers


class Custom404Test(unittest.TestCase):

  def setUp(self):
    self.services = service_manager.Services(
        project=fake.ProjectService())
    self.servlet = custom_404.ErrorPage(services=self.services)

  def testGatherPageData_NoProjectSpecified(self):
    """Project was not included in URL, so raise exception, will cause 400."""
    _, mr = testing_helpers.GetRequestObjects(
        path='/not/a/project/url')

    with self.assertRaises(exceptions.InputException):
      self.servlet.GatherPageData(mr)

  def testGatherPageData_Normal(self):
    """Return page_data dict with a 404 response code specified."""
    _project = self.services.project.TestAddProject('proj')
    _, mr = testing_helpers.GetRequestObjects(path='/p/proj/junk')

    page_data = self.servlet.GatherPageData(mr)
    self.assertEqual({'http_response_code': http_client.NOT_FOUND}, page_data)
