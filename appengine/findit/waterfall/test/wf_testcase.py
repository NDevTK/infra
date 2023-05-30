# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import copy
import mock

from google.appengine.api import users
from google.appengine.api import urlfetch

import gae_ts_mon

from gae_libs.testcase import TestCase
from model.wf_config import FinditConfig

_DEFAULT_CODE_COVERAGE_SETTINGS = {
    'serve_presubmit_coverage_data': True,
}

DEFAULT_CONFIG_DATA = {
    'code_coverage_settings': _DEFAULT_CODE_COVERAGE_SETTINGS,
}

SAMPLE_STEP_METADATA = {
    'waterfall_mastername': 'm',
    'waterfall_buildername': 'b',
    'canonical_step_name': 'browser_tests',
    'full_step_name': 'browser_tests on platform',
    'dimensions': {
        'os': 'platform'
    },
    'swarm_task_ids': ['1000', '1001']
}

SAMPLE_STEP_METADATA_NOT_SWARMED = {
    'waterfall_mastername': 'm',
    'waterfall_buildername': 'b',
    'canonical_step_name': 'browser_tests',
    'full_step_name': 'browser_tests on platform',
    'dimensions': {
        'os': 'platform'
    }
}


class WaterfallTestCase(TestCase):  # pragma: no cover.

  def UpdateUnitTestConfigSettings(self,
                                   config_property=None,
                                   override_data=None):
    """Sets up Findit's config for unit tests.

    Args:
      config_property: The name of the config property to update.
      override_data: A dict to override any default settings.
    """
    config_data = DEFAULT_CONFIG_DATA

    if config_property and override_data:
      config_data = copy.deepcopy(DEFAULT_CONFIG_DATA)
      config_data[config_property].update(override_data)

    FinditConfig.Get().Update(
        users.User(email='admin@chromium.org'), True, **config_data)

  def GetUnitTestConfigSettings(self):
    return FinditConfig.Get()

  def setUp(self):
    super(WaterfallTestCase, self).setUp()
    self.UpdateUnitTestConfigSettings()
    self.maxDiff = None
    # Make sure that no tests derived from this actually call urlfetch.fetch.
    mock.patch.object(
        urlfetch,
        'fetch',
        side_effect=AssertionError(
            'unittests must not perform actual network requests. Instead, '
            'mocks should be provided for the methods that do any network '
            'operations')).start()
    gae_ts_mon.reset_for_unittest(disable=True)

  def tearDown(self):
    mock.patch.stopall()
    super(WaterfallTestCase, self).tearDown()
