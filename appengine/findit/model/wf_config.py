# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Findit for Waterfall configuration."""

from google.appengine.ext import ndb

from gae_libs.model.versioned_config import VersionedConfig


class FinditConfig(VersionedConfig):
  """Global configuration of findit."""

  # A dict containing settings for Code Coverage. For example,
  # {
  #     'serve_presubmit_coverage_data': True,
  #     'project_banners': {
  #       'chromium/src': {
  #         'message':
  #           'browser_tests has been disabled. Coverage totals may be skewed.',
  #         'bug':
  #           937521,
  #       }
  #     },
  # }
  code_coverage_settings = ndb.JsonProperty(indexed=False, default={})

