# Copyright 2016 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Page class for generating somewhat informative project-page 404s.

This page class produces a mostly-empty project subpage, which helps
users find what they're looking for by providing navigational menus,
rather than telling them "404. That's an error. That's all we know."
which is maddeningly not helpful when we already have a project pb
loaded.
"""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

from six.moves import http_client
from framework import exceptions
from framework import flaskservlet


class ErrorPage(flaskservlet.FlaskServlet):
  """Page class for generating somewhat informative project-page 404s.

  This page class produces a mostly-empty project subpage, which helps
  users find what they're looking for by providing navigational menus,
  rather than telling them "404. That's an error. That's all we know."
  which is maddeningly not helpful when we already have a project pb
  loaded.
  """

  _PAGE_TEMPLATE = 'sitewide/project-404-page.ezt'

  def GatherPageData(self, mr):
    """Build up a dictionary of data values to use when rendering the page."""
    if not mr.project_name:
      raise exceptions.InputException('No project specified')
    return {
        'http_response_code': http_client.NOT_FOUND,
    }

  def Get404Page(self, **kwargs):
    return self.handler(**kwargs)
