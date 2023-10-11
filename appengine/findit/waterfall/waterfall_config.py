# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Determines support level for different steps for masters."""

from model.wf_config import FinditConfig


def GetCodeCoverageSettings():
  return FinditConfig().Get().code_coverage_settings
