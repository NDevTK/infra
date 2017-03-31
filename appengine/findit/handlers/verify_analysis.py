# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from gae_libs.handlers.base_handler import BaseHandler, Permission


class VerifyAnalysis(BaseHandler):
  PERMISSION_LEVEL = Permission.ANYONE

  def HandleGet(self):  # pylint: disable=R0201
    """Checks for revert or fix of CLs in a failed build cycle.

    Later it will be extended for coverage analysis.
    This corresponds to the "Verify after green" button in the following page:
    https://1-dot-findit-for-me.appspot.com/list-build?count=400&type=triage

    This endpoint serves JSON result.
    """
    return BaseHandler.CreateError('Not implemented yet!', 501)
