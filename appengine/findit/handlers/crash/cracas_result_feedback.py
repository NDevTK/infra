# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import urllib

from handlers.crash.result_feedback import ResultFeedback

_CRACAS_BASE_URL = 'https://crash.corp.google.com/browse'


class CracasResultFeedback(ResultFeedback):  # pragma: no cover

  @property
  def client(self):
    return 'cracas'
