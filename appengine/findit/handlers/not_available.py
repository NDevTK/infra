# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from gae_libs.handlers.base_handler import BaseHandler, Permission


class RedirectFlakePortal(BaseHandler):  # pragma: no cover.
  PERMISSION_LEVEL = Permission.ANYONE

  def HandleGet(self):
    return self.CreateRedirect('https://luci-analysis.appspot.com/')
