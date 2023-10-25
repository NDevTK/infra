# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import copy
from flask import Flask
import json
import mock
import re
import webtest

from google.appengine.api import users

from backend.handlers.update_component_config import (
    GetComponentClassifierConfig)
from backend.handlers.update_component_config import UpdateComponentConfig
from common.model.crash_config import CrashConfig
from frontend.handlers import crash_config
from gae_libs.http.http_client_appengine import HttpClientAppengine
from gae_libs.testcase import TestCase
from libs.http.retry_http_client import RetryHttpClient


_MOCK_OWNERS_MAPPINGS = json.dumps({
  'component-to-team': {
      'compoA': 'team1@chromium.org',
      'compoB': 'team2@chromium.org',
      'compoD': 'team4@chromium.org'
    },
  'dir-to-component': {
      'dirA': 'compoA',
      'dirB': 'compoB',
      'dirC': 'compoB',
      'dirE': 'compoE',
    }
})


_MOCK_CONFIG = {
    'component_info': [
        {
            'dirs': ['src/dirA'],
            'component': 'compoA',
            'team': 'team1@chromium.org'
        },
        {
            'dirs': ['src/dirB', 'src/dirC'],
            'component': 'compoB',
            'team': 'team2@chromium.org'
        },
        {
            'dirs': ['src/dirE'],
            'component': 'compoE'
        },
    ],
    'owner_mapping_url': 'url',
    'top_n': 4,
}


_MOCK_CURRENT_CONFIG = {
    'component_info': [
        {'dirs': ['src/dirA'], 'component': 'compoA',
         'team': 'team1@chromium.org'},
    ],

    'owner_mapping_url': 'url',
    'top_n': 4,
}


class DummyHttpClient(HttpClientAppengine):  # pragma: no cover.
  def __init__(self, config=None, response=None):
    super(DummyHttpClient, self).__init__()
    self.mock_owners_mappings = config or CrashConfig.Get().component_classifier
    self.response = response or _MOCK_OWNERS_MAPPINGS

  def Get(self, *_):  # pylint: disable=W
    if 'owner_mapping_url' in self.mock_owners_mappings:
      return 200, self.response, {}
    else:
      return 500, {}, {}


class UpdateComponentConfigTest(TestCase):
  """Tests utility functions and ``CrashConfig`` handler."""
  app_module = Flask(__name__)
  app_module.add_url_rule(
      '/process/update-component-config',
      view_func=UpdateComponentConfig().Handle,
      methods=['GET'])

  def setUp(self):
    super(UpdateComponentConfigTest, self).setUp()
    self.http_client_for_git = self.GetMockHttpClient()

  def testGetComponentClassifierConfig(self):
    component_classifier_config = GetComponentClassifierConfig(
        _MOCK_CURRENT_CONFIG, DummyHttpClient({'owner_mapping_url': 'url'}))
    expected_components = _MOCK_CONFIG['component_info']
    components = component_classifier_config['component_info']
    self.assertCountEqual(components, expected_components)

  def testGetComponentClassifierConfigNoOWNERS(self):
    component_classifier_config = GetComponentClassifierConfig(
        _MOCK_CURRENT_CONFIG, DummyHttpClient(config={'top_n': 3}))
    self.assertIsNone(component_classifier_config)

  @mock.patch(
      'backend.handlers.update_component_config.GetComponentClassifierConfig')
  def testHandleGet(self, mocked_get_component_classifier_config):
    mocked_get_component_classifier_config.return_value = _MOCK_CONFIG
    response = self.test_app.get('/process/update-component-config',
                                 headers={'X-AppEngine-Cron': 'true'})
    self.assertEqual(response.status_int, 200)
    self.assertDictEqual(_MOCK_CONFIG, CrashConfig.Get().component_classifier)
