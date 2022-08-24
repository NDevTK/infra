# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
""""Helper methods for structured logging."""

from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import google.cloud.logging

import settings


def log(struct):
  if settings.local_mode or settings.unit_test_mode:
    return

  logging_client = google.cloud.logging.Client()
  logger = logging_client.logger('python')
  logger.log_struct(struct)
