# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import datetime
import json
import mock
import re
import six

from flask import Flask
import webtest

from handlers import config
from model import wf_config
from testing_utils import testing
from google.appengine.api import users


_MOCK_CODE_COVERAGE_SETTINGS = {
    'serve_presubmit_coverage_data': True,
}

class ConfigTest(testing.AppengineTestCase):
  app_module = Flask(__name__)
  app_module.add_url_rule(
      '/waterfall/config',
      view_func=config.Configuration().Handle,
      methods=['GET', 'POST'])

  def testGetConfigurationSettings(self):
    config_data = {
        'code_coverage_settings': _MOCK_CODE_COVERAGE_SETTINGS,
    }

    self.mock_current_user(user_email='test@chromium.org', is_admin=True)

    wf_config.FinditConfig.Get().Update(
        users.GetCurrentUser(), True, message='message', **config_data)

    response = self.test_app.get('/waterfall/config', params={'format': 'json'})
    self.assertEquals(response.status_int, 200)

    expected_response = {
        'code_coverage_settings': _MOCK_CODE_COVERAGE_SETTINGS,
        'version': 1,
        'latest_version': 1,
        'updated_by': 'test',
        'updated_ts': response.json_body.get('updated_ts'),
        'message': 'message',
        'xsrf_token': response.json_body['xsrf_token'],
    }

    self.assertEquals(expected_response, response.json_body)

  def testGetVersionOfConfigurationSettings(self):
    self.mock_current_user(user_email='test@chromium.org', is_admin=True)

    config_data = {
        'code_coverage_settings': _MOCK_CODE_COVERAGE_SETTINGS,
    }
    wf_config.FinditConfig.Get().Update(
        users.GetCurrentUser(), True, message='message', **config_data)

    response = self.test_app.get(
        '/waterfall/config', params={
            'version': 1,
            'format': 'json'
        })
    self.assertEquals(response.status_int, 200)

    expected_response = {
        'code_coverage_settings': _MOCK_CODE_COVERAGE_SETTINGS,
        'version': 1,
        'latest_version': 1,
        'updated_by': 'test',
        'updated_ts': response.json_body.get('updated_ts'),
        'message': 'message',
        'xsrf_token': response.json_body['xsrf_token'],
    }

    self.assertEquals(expected_response, response.json_body)

  def testGetOutOfBoundsVersionOfConfigurationSettings(self):
    config_data = {
        'code_coverage_settings': _MOCK_CODE_COVERAGE_SETTINGS,
    }
    self.mock_current_user(user_email='test@chromium.org', is_admin=True)

    wf_config.FinditConfig.Get().Update(users.GetCurrentUser(), True,
                                        **config_data)

    self.assertRaisesRegexp(
        webtest.app.AppError,
        re.compile('The requested version is invalid or not found.',
                   re.MULTILINE | re.DOTALL),
        self.test_app.get,
        '/waterfall/config',
        params={
            'version': 0,
            'format': 'json'
        })
    self.assertRaisesRegexp(
        webtest.app.AppError,
        re.compile('The requested version is invalid or not found.',
                   re.MULTILINE | re.DOTALL),
        self.test_app.get,
        '/waterfall/config',
        params={
            'version': 2,
            'format': 'json'
        })

  def testFormatTimestamp(self):
    self.assertIsNone(config._FormatTimestamp(None))
    self.assertEqual(
        '2016-02-25 01:02:03',
        config._FormatTimestamp(
            datetime.datetime(2016, 2, 25, 1, 2, 3, 123456)))

  @mock.patch('common.token.ValidateAuthToken')
  def testPostConfigurationSettings(self, mocked_ValidateAuthToken):
    self.mock_current_user(user_email='test@chromium.org', is_admin=True)
    mocked_ValidateAuthToken.side_effect = [(True, False)]

    params = {
        'format':
            'json',
        'code_coverage_settings':
            json.dumps(_MOCK_CODE_COVERAGE_SETTINGS),
        'message':
            'reason',
    }

    response = self.test_app.post('/waterfall/config', params=params)

    expected_response = {
        'code_coverage_settings': _MOCK_CODE_COVERAGE_SETTINGS,
        'version': 1,
        'latest_version': 1,
        'updated_by': 'test',
        'updated_ts': response.json_body.get('updated_ts'),
        'message': 'reason',
        'xsrf_token': response.json_body['xsrf_token'],
    }

    self.assertEquals(expected_response, response.json_body)

  @mock.patch('common.token.ValidateAuthToken')
  def testPostConfigurationSettingsFail(self, mocked_ValidateAuthToken):
    self.mock_current_user(user_email='test@chromium.org', is_admin=True)
    mocked_ValidateAuthToken.side_effect = [(True, False)]

    self.assertRaisesRegexp(
        webtest.app.AppError,
        re.compile('not present in config', re.MULTILINE | re.DOTALL),
        self.test_app.post,
        '/waterfall/config',
        params={
            'format': 'json',
            'message': 'forgot how to config',
        })

  def testValidateConfig(self):
    spec = {
        'required_int':
            int,
        'required_float': (float,),  # One item tuple must be supported.
        'optional_string': (str, False),
        'required_even_length_list': (list, True,
                                      lambda x: ['odd'] * (len(x) % 2)),
        'optional_nested_dict': (dict, False, {
            'inner_key': int
        })
    }
    good_config = {
        'required_int': 1,
        'required_float': 2.5,
        'optional_string': 'hello world',
        'required_even_length_list': [1, 2, 3, 4],
        'optional_nested_dict': {
            'inner_key': 5,
        }
    }
    good_config_2 = {
        'required_int': 1,
        'required_float': 2,  # int should satisfy float req.
        'optional_string': u'hello world',  # Unicode should satisfy str req.
        'required_even_length_list': [],
        'optional_nested_dict': {
            'inner_key': 5,
            'extra_key': 'whimsy',  # extra keys should not break the config.
        }
    }
    good_config_3 = {
        'required_int': 1,
        'required_float': 2.5,  # int should satisfy float req.
        'required_even_length_list': [],
        # optional should be optional.
    }
    bad_types = {
        'required_int': 1.5,
        'required_float': False,
        'optional_string': ['h', 'e', 'l', 'l', 'o'],
        'required_even_length_list': (1, 2, 3, 4),
        'optional_nested_dict': {
            'inner_key': 5.0,
        }
    }
    bad_custom_validation = {
        'required_int': 1,
        'required_float': 2.5,
        'optional_string': 'hello world',
        'required_even_length_list': [1, 2, 3],
        'optional_nested_dict': {
            'inner_missing': 5,
        }
    }
    self.assertEqual([], config._ValidateConfig('', good_config, spec))
    self.assertEqual([], config._ValidateConfig('', good_config_2, spec))
    self.assertEqual([], config._ValidateConfig('', good_config_3, spec))

    if six.PY2:
      self.assertEqual([
          'Expected key inner_key, value 5.0 to be <type \'int\'> in config '
          '/optional_nested_dict', 'Expected key optional_string, value '
          '[\'h\', \'e\', \'l\', \'l\', \'o\''
          '] to be (<type \'basestring\'>,) in config ',
          'Expected key required_int, value 1.5 to be '
          '<type \'int\'> in config ',
          'Expected key required_even_length_list, value (1, 2, 3, 4) to be '
          '<type \'list\'> in config '
      ], config._ValidateConfig('', bad_types, spec))
      self.assertEqual([
          'Required key inner_key not present in config /optional_nested_dict',
          'odd'
      ], config._ValidateConfig('', bad_custom_validation, spec))
    else:
      self.assertCountEqual([
          'Expected key inner_key, value 5.0 to be <class \'int\'> in config '
          '/optional_nested_dict', 'Expected key optional_string, value '
          '[\'h\', \'e\', \'l\', \'l\', \'o\''
          '] to be (<class \'str\'>,) in config ',
          'Expected key required_int, value 1.5 to be '
          '<class \'int\'> in config ',
          'Expected key required_even_length_list, value (1, 2, 3, 4) to be '
          '<class \'list\'> in config '
      ], config._ValidateConfig('', bad_types, spec))
      self.assertCountEqual([
          'Required key inner_key not present in config /optional_nested_dict',
          'odd'
      ], config._ValidateConfig('', bad_custom_validation, spec))
    self.assertNotEqual([], config._ValidateConfig('', 'not_a_dict', spec))
