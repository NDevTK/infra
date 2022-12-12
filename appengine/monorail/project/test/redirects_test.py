# Copyright 2016 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Tests for project handlers that redirect."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

from six.moves import http_client
import unittest

import flask

from framework import urls
from project import redirects
from services import service_manager
from testing import fake
from testing import testing_helpers


class WikiRedirectTest(unittest.TestCase):

  def setUp(self):
    self.services = service_manager.Services()
    self.servlet = redirects.WikiRedirect(services=self.services)
    self.servlet.response = flask.Response()
    self.project = fake.Project()
    self.servlet.mr = testing_helpers.MakeMonorailRequest(
        project=self.project)

  def testRedirect_NoSuchProject(self):
    """Visiting a project that we don't host is 404."""
    self.servlet.mr.project = None
    self.servlet.get()
    self.assertEqual(http_client.NOT_FOUND, self.servlet.response.status_code)

  def testRedirect_NoDocsSpecified(self):
    """Visiting any old wiki URL goes to admin intro by default."""
    self.servlet.get()
    self.assertEqual(
        http_client.MOVED_PERMANENTLY, self.servlet.response.status_code)
    self.assertTrue(
        self.servlet.response.location.endswith(urls.ADMIN_INTRO))

  def testRedirect_DocsSpecified(self):
    """Visiting any old wiki URL goes to project docs URL."""
    self.project.docs_url = 'some_url'
    self.servlet.get()
    self.assertEqual(
        http_client.MOVED_PERMANENTLY, self.servlet.response.status_code)
    self.assertEqual('some_url', self.servlet.response.location)


class SourceRedirectTest(unittest.TestCase):

  def setUp(self):
    self.services = service_manager.Services()
    self.servlet = redirects.SourceRedirect(services=self.services)
    self.servlet.response = flask.Response()
    self.project = fake.Project()
    self.servlet.mr = testing_helpers.MakeMonorailRequest(
        project=self.project)

  def testRedirect_NoSuchProject(self):
    """Visiting a project that we don't host is 404."""
    self.servlet.mr.project = None
    self.servlet.get()
    self.assertEqual(http_client.NOT_FOUND, self.servlet.response.status_code)

  def testRedirect_NoSrcSpecified(self):
    """Visiting any old source code URL goes to admin intro by default."""
    self.servlet.get()
    self.assertEqual(
        http_client.MOVED_PERMANENTLY, self.servlet.response.status_code)
    self.assertTrue(
        self.servlet.response.location.endswith(urls.ADMIN_INTRO))

  def testRedirect_SrcSpecified(self):
    """Visiting any old source code URL goes to project source URL."""
    self.project.source_url = 'some_url'
    self.servlet.get()
    self.assertEqual(
        http_client.MOVED_PERMANENTLY, self.servlet.response.status_code)
    self.assertEqual('some_url', self.servlet.response.location)
