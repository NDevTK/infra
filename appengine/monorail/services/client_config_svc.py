# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import base64
import json
import logging
import os
import time
from six.moves import urllib
import webapp2
import flask

from google.appengine.api import app_identity
from google.appengine.api import urlfetch
from google.appengine.ext import db
from google.protobuf import text_format

from infra_libs import ts_mon

import settings
from framework import framework_constants
from proto import api_clients_config_pb2


CONFIG_FILE_PATH = os.path.join(
    os.path.dirname(os.path.dirname(os.path.realpath(__file__))),
    'testing', 'api_clients.cfg')
LUCI_CONFIG_URL = (
    'https://luci-config.appspot.com/_ah/api/config/v1/config_sets'
    '/services/monorail-prod/config/api_clients.cfg')


client_config_svc = None
service_account_map = None
qpm_dict = None
allowed_origins_set = None


class ClientConfig(db.Model):
  configs = db.TextProperty()


_CONFIG_LOADS = ts_mon.CounterMetric(
    'monorail/client_config_svc/loads', 'Results of fetches from luci-config.',
    [ts_mon.BooleanField('success'),
     ts_mon.StringField('type')])


def _process_response(response):
  try:
    content = json.loads(response.content)
  except ValueError:
    logging.error('Response was not JSON: %r', response.content)
    _CONFIG_LOADS.increment({'success': False, 'type': 'json-load-error'})
    raise

  try:
    config_content = content['content']
  except KeyError:
    logging.error('JSON contained no content: %r', content)
    _CONFIG_LOADS.increment({'success': False, 'type': 'json-key-error'})
    raise

  try:
    content_text = base64.b64decode(config_content)
  except TypeError:
    logging.error('Content was not b64: %r', config_content)
    _CONFIG_LOADS.increment({'success': False, 'type': 'b64-decode-error'})
    raise

  try:
    cfg = api_clients_config_pb2.ClientCfg()
    text_format.Merge(content_text, cfg)
  except:
    logging.error('Content was not a valid ClientCfg proto: %r', content_text)
    _CONFIG_LOADS.increment({'success': False, 'type': 'proto-load-error'})
    raise

  return content_text


def GetLoadApiClientConfigs():
  global service_account_map
  global qpm_dict
  authorization_token, _ = app_identity.get_access_token(
      framework_constants.OAUTH_SCOPE)
  response = urlfetch.fetch(
      LUCI_CONFIG_URL,
      method=urlfetch.GET,
      follow_redirects=False,
      headers={
          'Content-Type': 'application/json; charset=UTF-8',
          'Authorization': 'Bearer ' + authorization_token
      })

  if response.status_code != 200:
    logging.error('Invalid response from luci-config: %r', response)
    _CONFIG_LOADS.increment({'success': False, 'type': 'luci-cfg-error'})
    flask.abort(500, 'Invalid response from luci-config')

  try:
    content_text = _process_response(response)
  except Exception as e:
    flask.abort(500, str(e))

  logging.info('luci-config content decoded: %r.', content_text)
  configs = ClientConfig(configs=content_text, key_name='api_client_configs')
  configs.put()
  service_account_map = None
  qpm_dict = None
  _CONFIG_LOADS.increment({'success': True, 'type': 'success'})

  return ''


class ClientConfigService(object):
  """The persistence layer for client config data."""

  # Reload no more than once every 15 minutes.
  # Different GAE instances can load it at different times,
  # so clients may get inconsistence responses shortly after allowlisting.
  EXPIRES_IN = 15 * framework_constants.SECS_PER_MINUTE

  def __init__(self):
    self.client_configs = None
    self.load_time = 0

  def GetConfigs(self, use_cache=True, cur_time=None):
    """Read client configs."""

    cur_time = cur_time or int(time.time())
    force_load = False
    if not self.client_configs:
      force_load = True
    elif not use_cache:
      force_load = True
    elif cur_time - self.load_time > self.EXPIRES_IN:
      force_load = True

    if force_load:
      if settings.local_mode or settings.unit_test_mode:
        self._ReadFromFilesystem()
      else:
        self._ReadFromDatastore()

    return self.client_configs

  def _ReadFromFilesystem(self):
    try:
      with open(CONFIG_FILE_PATH, 'r') as f:
        content_text = f.read()
      logging.info('Read client configs from local file.')
      cfg = api_clients_config_pb2.ClientCfg()
      text_format.Merge(content_text, cfg)
      self.client_configs = cfg
      self.load_time = int(time.time())
    except Exception as e:
      logging.exception('Failed to read client configs: %s', e)

  def _ReadFromDatastore(self):
    entity = ClientConfig.get_by_key_name('api_client_configs')
    if entity:
      cfg = api_clients_config_pb2.ClientCfg()
      text_format.Merge(entity.configs, cfg)
      self.client_configs = cfg
      self.load_time = int(time.time())
    else:
      logging.error('Failed to get api client configs from datastore.')

  def GetClientIDEmails(self):
    """Get client IDs and Emails."""
    self.GetConfigs(use_cache=True)
    client_ids = [c.client_id for c in self.client_configs.clients]
    client_emails = [c.client_email for c in self.client_configs.clients]
    return client_ids, client_emails

  def GetDisplayNames(self):
    """Get client display names."""
    self.GetConfigs(use_cache=True)
    names_dict = {}
    for client in self.client_configs.clients:
      if client.display_name:
        names_dict[client.client_email] = client.display_name
    return names_dict

  def GetQPM(self):
    """Get client qpm limit."""
    self.GetConfigs(use_cache=True)
    qpm_map = {}
    for client in self.client_configs.clients:
      if client.HasField('qpm_limit'):
        qpm_map[client.client_email] = client.qpm_limit
    return qpm_map

  def GetAllowedOriginsSet(self):
    """Get the set of all allowed origins."""
    self.GetConfigs(use_cache=True)
    origins = set()
    for client in self.client_configs.clients:
      origins.update(client.allowed_origins)
    return origins


def GetClientConfigSvc():
  global client_config_svc
  if client_config_svc is None:
    client_config_svc = ClientConfigService()
  return client_config_svc


def GetServiceAccountMap():
  # typ: () -> Mapping[str, str]
  """Returns only service accounts that have specified display_names."""
  global service_account_map
  if service_account_map is None:
    service_account_map = GetClientConfigSvc().GetDisplayNames()
  return service_account_map


def GetQPMDict():
  global qpm_dict
  if qpm_dict is None:
    qpm_dict = GetClientConfigSvc().GetQPM()
  return qpm_dict


def GetAllowedOriginsSet():
  global allowed_origins_set
  if allowed_origins_set is None:
    allowed_origins_set = GetClientConfigSvc().GetAllowedOriginsSet()
  return allowed_origins_set
