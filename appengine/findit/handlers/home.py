# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from common.base_handler import BaseHandler, Permission


class Home(BaseHandler):
  PERMISSION_LEVEL = Permission.ANYONE

  def HandleGet(self, **kwargs):
    """Renders home pages."""
    return {
        'template': 'home.html',
    }
