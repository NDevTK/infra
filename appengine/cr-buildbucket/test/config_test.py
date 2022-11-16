# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import copy
import hashlib
import logging

from parameterized import parameterized

from components import utils

utils.fix_protobuf_package()

from google.appengine.api import memcache
from google.appengine.ext import ndb
from google.protobuf import text_format

from components import config as config_component
from components.config import validation_context
from testing_utils import testing
import mock

from go.chromium.org.luci.buildbucket.proto import build_pb2
from go.chromium.org.luci.buildbucket.proto import builder_common_pb2
from go.chromium.org.luci.buildbucket.proto import project_config_pb2
from go.chromium.org.luci.buildbucket.proto import service_config_pb2
from test import test_util
import config
import errors


def short_bucket_cfg(cfg):
  cfg = copy.deepcopy(cfg)
  cfg.name = config.short_bucket_name(cfg.name)
  return cfg


def without_builders(cfg):
  cfg = copy.deepcopy(cfg)
  if config.is_swarming_config(cfg):
    del cfg.swarming.builders[:]
  return cfg


# NOTE: This has the pool dimension, but other uses of LUCI_CHROMIUM_TRY
# do NOT list pool dimension explicitly; This is because config.py has api
# little-used feature where pool is 'automatically' filled in from the project
# and bucket name as 'luci.$project.$bucket' if it's omitted.
LUCI_CHROMIUM_TRY = test_util.parse_bucket_cfg(
    '''
    name: "luci.chromium.try"
    swarming {
      task_template_canary_percentage { value: 10 }
      builders {
        name: "linux"
        swarming_host: "swarming.example.com"
        task_template_canary_percentage { value: 10 }
        dimensions: "os:Linux"
        dimensions: "pool:luci.chromium.try"
        recipe {
          cipd_package: "infra/recipe_bundle"
          cipd_version: "refs/heads/master"
          name: "x"
        }
      }
    }
'''
)

LUCI_DART_TRY = test_util.parse_bucket_cfg(
    '''
    name: "luci.dart.try"
    swarming {
      builders {
        name: "linux"
        dimensions: "pool:Dart.LUCI"
        recipe {
          cipd_package: "infra/recipe_bundle"
          cipd_version: "refs/heads/master"
          name: "x"
        }
      }
    }
    '''
)

MASTER_TRYSERVER_CHROMIUM_LINUX = test_util.parse_bucket_cfg(
    '''
    name: "master.tryserver.chromium.linux"
    '''
)

MASTER_TRYSERVER_CHROMIUM_WIN = test_util.parse_bucket_cfg(
    '''
    name: "master.tryserver.chromium.win"
    '''
)

MASTER_TRYSERVER_CHROMIUM_MAC = test_util.parse_bucket_cfg(
    '''
    name: "master.tryserver.chromium.mac"
    '''
)

MASTER_TRYSERVER_V8 = test_util.parse_bucket_cfg(
    '''
    name: "master.tryserver.v8"
    '''
)


def parse_cfg(text):
  cfg = project_config_pb2.BuildbucketCfg()
  text_format.Merge(text, cfg)
  return cfg


def errmsg(text):
  return validation_context.Message(
      severity=logging.ERROR, text=text.decode('utf-8')
  )


class ConfigTest(testing.AppengineTestCase):

  def test_get_bucket(self):
    config.put_bucket('chromium', 'deadbeef', LUCI_CHROMIUM_TRY)
    rev, cfg = config.get_bucket('chromium/try')
    self.assertEqual(rev, 'deadbeef')
    self.assertEqual(cfg, without_builders(short_bucket_cfg(LUCI_CHROMIUM_TRY)))

    self.assertIsNone(config.get_bucket('chromium/nonexistent')[0])

  def test_get_buckets_async(self):
    config.put_bucket('chromium', 'deadbeef', MASTER_TRYSERVER_CHROMIUM_LINUX)
    config.put_builders(
        'chromium', 'master.tryserver.chromium.linux',
        *MASTER_TRYSERVER_CHROMIUM_LINUX.swarming.builders
    )
    config.put_bucket('chromium', 'deadbeef', LUCI_CHROMIUM_TRY)
    config.put_builders('chromium', 'try', *LUCI_CHROMIUM_TRY.swarming.builders)
    config.put_bucket('dart', 'deadbeef', LUCI_DART_TRY)
    config.put_builders('dart', 'try', *LUCI_DART_TRY.swarming.builders)

    # Without builders.
    actual = config.get_buckets_async().get_result()
    expected = {
        'chromium/master.tryserver.chromium.linux':
            without_builders(MASTER_TRYSERVER_CHROMIUM_LINUX),
        'chromium/try':
            without_builders(short_bucket_cfg(LUCI_CHROMIUM_TRY)),
        'dart/try':
            without_builders(short_bucket_cfg(LUCI_DART_TRY)),
    }
    self.assertEqual(actual, expected)

    # With builders.
    actual = config.get_buckets_async(include_builders=True).get_result()
    expected = {
        'chromium/master.tryserver.chromium.linux':
            MASTER_TRYSERVER_CHROMIUM_LINUX,
        'chromium/try':
            short_bucket_cfg(LUCI_CHROMIUM_TRY),
        'dart/try':
            short_bucket_cfg(LUCI_DART_TRY),
    }
    self.assertEqual(actual, expected)

  def test_get_buckets_async_with_bucket_ids(self):
    config.put_bucket('chromium', 'deadbeef', LUCI_CHROMIUM_TRY)
    config.put_builders('chromium', 'try', *LUCI_CHROMIUM_TRY.swarming.builders)
    config.put_bucket('chromium', 'deadbeef', MASTER_TRYSERVER_CHROMIUM_WIN)
    config.put_builders(
        'chromium', 'master.tryserver.chromium.win',
        *MASTER_TRYSERVER_CHROMIUM_WIN.swarming.builders
    )

    # Without builders.
    actual = config.get_buckets_async(bucket_ids=['chromium/try']).get_result()
    expected = {
        'chromium/try': without_builders(short_bucket_cfg(LUCI_CHROMIUM_TRY))
    }
    self.assertEqual(actual, expected)

    # With builders.
    actual = config.get_buckets_async(
        bucket_ids=['chromium/master.tryserver.chromium.win'],
        include_builders=True
    ).get_result()
    expected = {
        'chromium/master.tryserver.chromium.win': MASTER_TRYSERVER_CHROMIUM_WIN
    }
    self.assertEqual(actual, expected)

  def test_get_buckets_async_with_bucket_ids_not_found(self):
    bid = 'chromium/try'
    actual = config.get_buckets_async([bid]).get_result()
    self.assertEqual(actual, {bid: None})

  def test_get_all_bucket_ids_cold_cache(self):
    config.put_bucket(
        'chromium', 'deadbeef', test_util.parse_bucket_cfg('name: "try"')
    )
    config.put_bucket(
        'chromium', 'deadbeef', test_util.parse_bucket_cfg('name: "ci"')
    )
    config.put_bucket(
        'v8', 'deadbeef', test_util.parse_bucket_cfg('name: "ci"')
    )
    ids = config.get_all_bucket_ids_async().get_result()
    self.assertEqual(ids, ['chromium/ci', 'chromium/try', 'v8/ci'])
    # for coverage test which covers the line #274 in config.py to read from
    # memecache directly.
    ids = config.get_all_bucket_ids_async().get_result()
    self.assertEqual(ids, ['chromium/ci', 'chromium/try', 'v8/ci'])

  def resolve_bucket(self, bucket_name):
    return config.resolve_bucket_name_async(bucket_name).get_result()

  def test_resolve_bucket_name_async_does_not_exist(self):
    self.assertIsNone(self.resolve_bucket('try'))

  def test_resolve_bucket_name_async_unique(self):
    config.put_bucket('chromium', 'deadbeef', LUCI_CHROMIUM_TRY)
    self.assertEqual(self.resolve_bucket('try'), 'chromium/try')

  def test_resolve_bucket_name_async_ambiguous(self):
    config.put_bucket('chromium', 'deadbeef', LUCI_CHROMIUM_TRY)
    config.put_bucket('dart', 'deadbeef', LUCI_DART_TRY)
    with self.assertRaisesRegexp(errors.InvalidInputError, r'ambiguous'):
      self.resolve_bucket('try')

  def test_resolve_bucket_name_async_cache_key(self):
    config.put_bucket('chromium', 'deadbeef', LUCI_CHROMIUM_TRY)
    config.put_bucket('chromium', 'deadbeef', MASTER_TRYSERVER_CHROMIUM_LINUX)
    self.assertEqual(self.resolve_bucket('try'), 'chromium/try')
    self.assertEqual(
        self.resolve_bucket('master.tryserver.chromium.linux'),
        'chromium/master.tryserver.chromium.linux'
    )

  def cfg_validation_test(self, cfg, expected_messages):
    ctx = config_component.validation.Context()
    ctx.config_set = 'projects/chromium'
    config.validate_buildbucket_cfg(cfg, ctx)
    self.assertEqual(expected_messages, ctx.result().messages)

  def test_validate_buildbucket_cfg_success(self):
    self.cfg_validation_test(
        parse_cfg(
            '''
      buckets {
        name: "good.name"
      }
      buckets {
        name: "good.name2"
      }
      '''
        ), []
    )

  def test_validate_buildbucket_cfg_fail(self):
    self.cfg_validation_test(
        parse_cfg(
            '''
      buckets {
        name: "a"
      }
      buckets {
        name: "a"
      }
      buckets {}
      buckets { name: "luci.x" }
      '''
        ), [
            errmsg('Bucket a: duplicate bucket name'),
            errmsg('Bucket #3: invalid name: Bucket not specified'),
            errmsg(
                'Bucket luci.x: invalid name: Bucket must start with '
                '"luci.chromium." because it starts with "luci." and is defined'
                ' in the chromium project'
            ),
        ]
    )

  def test_validate_buildbucket_cfg_unsorted(self):
    self.cfg_validation_test(
        parse_cfg(
            '''
            buckets { name: "c" }
            buckets { name: "b" }
            buckets { name: "a" }
            '''
        ),
        [
            validation_context.Message(
                severity=logging.WARNING,
                text='Bucket b: out of order',
            ),
            validation_context.Message(
                severity=logging.WARNING,
                text='Bucket a: out of order',
            ),
        ],
    )

  @mock.patch('components.config.get_config_set_location', autospec=True)
  def test_get_buildbucket_cfg_url(self, get_config_set_location):
    get_config_set_location.return_value = (
        'https://chromium.googlesource.com/chromium/src/+/infra/config'
    )

    url = config.get_buildbucket_cfg_url('chromium')
    self.assertEqual(
        url, (
            'https://chromium.googlesource.com/chromium/src/+/'
            'refs/heads/infra/config/testbed-test.cfg'
        )
    )


class ValidateBucketIDTest(testing.AppengineTestCase):

  @parameterized.expand([
      ('chromium/try',),
      ('chrome-internal/try',),
      ('chrome-internal/try.x',),
  ])
  def test_valid(self, bucket_id):
    config.validate_bucket_id(bucket_id)

  @parameterized.expand([
      ('a/b/c',),
      ('a b/c',),
      ('chromium/luci.chromium.try',),
  ])
  def test_invalid(self, bucket_id):
    with self.assertRaises(errors.InvalidInputError):
      config.validate_bucket_id(bucket_id)

  @parameterized.expand([
      ('chromium',),
      ('a:b',),
  ])
  def test_assertions(self, bucket_id):
    # We must never pass legacy bucket id to validate_bucket_id().
    with self.assertRaises(AssertionError):
      config.validate_bucket_id(bucket_id)


class BuilderMatchesTest(testing.AppengineTestCase):

  @parameterized.expand([
      ([], [], True),
      ([], ['chromium/.+'], False),
      ([], ['v8/.+'], True),
      (['chromium/.+'], [], True),
      (['v8/.+'], [], False),
  ])
  def test_builder_matches(self, regex, regex_exclude, expected):
    predicate = service_config_pb2.BuilderPredicate(
        regex=regex, regex_exclude=regex_exclude
    )
    builder_id = builder_common_pb2.BuilderID(
        project='chromium',
        bucket='try',
        builder='linux-rel',
    )
    actual = config.builder_matches(builder_id, predicate)
    self.assertEqual(expected, actual)
