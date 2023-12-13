# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Handles requests to the findit config page."""

import logging
import json

from common import token
from common.base_handler import BaseHandler, Permission
from model import wf_config
from waterfall import waterfall_config

from google.appengine.api import users


def _ValidateConfig(name, d, spec):
  """Validate that a given config matches the specification.

  Configs are dicts, and specs are dicts in the following format:
  {
    # Either format is okay. required is a boolean, defaults to True if not
    # given.
    'key_name': type,
    'key_name': (type, required),
    'key_name': (type, required, validator_or_nested_spec),
   }

  This function iterates over every key in the spec and
    - makes sure that the key is present in the given config(d) if required is
      true,
    - makes sure that the value is the type(s) given in the spec, note that it
      is okay to pass a tuple of types if the spec itself is a tuple
        e.g.((int, float), True)
    - makes sure that the value passes a custom validation if custom validation,
      if a validator function is provided, or if a nested spec is provided, it
      recursively calls _ValidateConfig on the value of the key (such as a
      nested dict)

  This function returns a list of errors (strings). It is expected that any
  custom validation functions will return a list of errors.

  A return value of [], indicates that there are no errors."""

  errors = []

  if not isinstance(d, dict):
    err = 'Expect %s to be a dictionary in config %s' % (d, name)
    logging.error(err)
    return [err]

  for key in spec:
    requirements = spec[key]

    # Sane defaults.
    required_type = int
    required_key = True
    custom_validator = None

    if isinstance(requirements, tuple):
      tuple_length = len(requirements)
      if tuple_length == 1:
        required_type = requirements[0]
      elif tuple_length == 2:
        required_type, required_key = requirements
      else:
        assert tuple_length == 3, 'The config tuple length must be < 3'
        required_type, required_key, custom_validator = requirements
    else:
      required_type = requirements

    if required_type == float:
      required_type = (int, float)
    elif required_type == str:
      required_type = basestring

    # Actual validation.

    # Validate key presence.
    if required_key and not key in d:
      err = 'Required key %s not present in config %s' % (key, name)
      logging.error(err)
      errors.append(err)
    # Validate type.
    elif key in d and not isinstance(d[key], required_type):
      err = 'Expected key %s, value %r to be %s in config % s' % (
          key, d[key], required_type, name)
      logging.error(err)
      errors.append(err)
    # Custom validator is a spec.
    elif (key in d and isinstance(custom_validator, dict)):
      errors += _ValidateConfig('%s/%s' % (name, key), d[key], custom_validator)
    # Custom validator is a function.
    elif key in d and callable(custom_validator):
      inner_errors = custom_validator(d[key])
      errors += inner_errors
      err = 'Key %s, value %r in config %s failed: %s' % (key, d[key], name,
                                                          inner_errors)
      logging.error(err)

  return errors


# Maps config properties to their validation specs.
# Please keep this config sorted by key name.
#
# Configs are dicts, and specs are dicts in the following format:
# {
#   'key_name': type,   # This implies required = True, and no custom validator.
#   'key_name': (type, required),
#   'key_name': (type, required, validator_or_nested_spec),
#  }
_CONFIG_SPEC = {  # yapf: disable
    'code_coverage_settings': (dict, True, {
        'serve_presubmit_coverage_data': bool,
    }),
}


def _FormatTimestamp(timestamp):
  if not timestamp:
    return None
  return timestamp.strftime('%Y-%m-%d %H:%M:%S')


class Configuration(BaseHandler):
  PERMISSION_LEVEL = Permission.ADMIN

  @token.AddXSRFToken(action_id='config')
  def HandleGet(self, **kwargs):
    version = self.request.values.get('version')
    if version is not None:
      version = int(version)
    settings = wf_config.FinditConfig.Get(version)
    if not settings:
      return BaseHandler.CreateError(
          'The requested version is invalid or not found.', 400)

    latest_version = settings.GetLatestVersionNumber()

    data = {
        'code_coverage_settings': settings.code_coverage_settings,
        'version': settings.version_number,
        'latest_version': latest_version,
        'updated_by': settings.updated_by,
        'updated_ts': _FormatTimestamp(settings.updated_ts),
        'message': settings.message,
    }

    return {'template': 'config.html', 'data': data}

  @token.VerifyXSRFToken(action_id='config')
  def HandlePost(self, **kwargs):
    new_config_dict = {}
    for name in self.request.values.keys():
      if name not in ('format', 'xsrf_token', 'message'):
        new_config_dict[name] = json.loads(self.request.values[name])

    message = self.request.values.get('message')
    if not message:  # pragma: no cover
      return self.CreateError('Please provide the reason to update the config',
                              400)

    errors = _ValidateConfig('', new_config_dict, _CONFIG_SPEC)
    if errors:
      return BaseHandler.CreateError(
          'New configuration settings is not properly formatted.\n'
          'The following errors were detected \n %s' % '\n'.join(errors), 400)

    wf_config.FinditConfig.Get().Update(
        users.get_current_user(),
        users.IsCurrentUserAdmin(),
        message=message,
        **new_config_dict)

    return self.HandleGet(**kwargs)
