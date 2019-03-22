# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import base64
import contextlib
import copy
import datetime
import json
import logging

from parameterized import parameterized

from components import utils
utils.fix_protobuf_package()

from google import protobuf
from google.appengine.ext import ndb
from google.protobuf import timestamp_pb2

from components import auth
from components import net
from components import utils
from testing_utils import testing
from webob import exc
import mock
import webapp2

from legacy import api_common
from proto import build_pb2
from proto import common_pb2
from proto import launcher_pb2
from proto.config import project_config_pb2
from proto.config import service_config_pb2
from test import test_util
from test.test_util import future, future_exception
import bbutil
import errors
import isolate
import model
import swarming
import user

linux_CACHE_NAME = (
    'builder_ccadafffd20293e0378d1f94d214c63a0f8342d1161454ef0acfa0405178106b'
)

NOW = datetime.datetime(2015, 11, 30)


def tspb(seconds, nanos=0):
  return timestamp_pb2.Timestamp(seconds=seconds, nanos=nanos)


class BaseTest(testing.AppengineTestCase):
  maxDiff = None

  def setUp(self):
    super(BaseTest, self).setUp()
    user.clear_request_cache()

    self.patch('tq.enqueue_async', autospec=True, return_value=future(None))

    self.now = NOW
    self.patch(
        'components.utils.utcnow', autospec=True, side_effect=lambda: self.now
    )

    self.settings = service_config_pb2.SettingsCfg(
        swarming=service_config_pb2.SwarmingSettings(
            milo_hostname='milo.example.com',
        ),
    )
    self.patch(
        'config.get_settings_async',
        autospec=True,
        return_value=future(self.settings)
    )

    builder_cfg_text = '''
      name: "linux"
      swarming_host: "swarming.example.com"
      swarming_tags: "buildertag:yes"
      swarming_tags: "commontag:yes"
      dimensions: "cores:8"
      dimensions: "os:Ubuntu"
      dimensions: "pool:Chrome"
      priority: 108
      build_numbers: YES
      service_account: "robot@example.com"
      recipe {
        name: "recipe"
        cipd_package: "infra/recipe_bundle"
        cipd_version: "refs/heads/master"
        properties_j: "predefined-property:\\\"x\\\""
        properties_j: "predefined-property-bool:true"
      }
      caches {
        path: "a"
        name: "a"
      }
      caches {
        path: "git_cache"
        name: "git_chromium"
      }
      caches {
        path: "out"
        name: "build_chromium"
      }
    '''
    self.builder_cfg = project_config_pb2.Builder()
    protobuf.text_format.Merge(builder_cfg_text, self.builder_cfg)


class TaskDefTest(BaseTest):

  def setUp(self):
    super(TaskDefTest, self).setUp()

    self.task_template = {
        'name':
            'bb-${build_id}-${project}-${builder}',
        'priority':
            '100',
        'tags': [
            (
                'log_location:logdog://luci-logdog-dev.appspot.com/${project}/'
                'buildbucket/${hostname}/${build_id}/+/annotations'
            ),
            'luci_project:${project}',
        ],
        'task_slices': [{
            'expiration_secs': '3600',
            'properties': {
                'execution_timeout_secs':
                    '3600',
                'extra_args': [
                    'cook',
                    '-recipe',
                    '${recipe}',
                    '-properties',
                    '${properties_json}',
                    '-logdog-project',
                    '${project}',
                ],
                'caches': [{
                    'path': '${cache_dir}/builder',
                    'name': 'builder_${builder_hash}',
                    'wait_for_warm_cache_secs': 60,
                }],
                'cipd_input': {
                    'packages': [
                        {
                            'package_name': 'infra/test/bar/${os_ver}',
                            'path': '.',
                            'version': 'latest',
                        },
                        {
                            'package_name': 'infra/test/foo/${platform}',
                            'path': 'third_party',
                            'version': 'stable',
                        },
                    ],
                },
            },
            'wait_for_capacity': False,
        },],
        'numerical_value_for_coverage_in_format_obj':
            42,
    }
    self.task_template_canary = self.task_template.copy()
    self.task_template_canary['name'] += '-canary'

    def get_self_config_async(path, *_args, **_kwargs):
      if path not in ('swarming_task_template.json',
                      'swarming_task_template_canary.json'):  # pragma: no cover
        self.fail()

      if path == 'swarming_task_template.json':
        template = self.task_template
      else:
        template = self.task_template_canary
      return future((
          'template_rev', json.dumps(template) if template is not None else None
      ))

    self.patch(
        'components.config.get_self_config_async',
        side_effect=get_self_config_async
    )

    self.patch(
        'google.appengine.api.app_identity.get_default_version_hostname',
        return_value='cr-buildbucket.appspot.com'
    )

  def prepare_task_def(self, build):
    return swarming.prepare_task_def_async(build, self.builder_cfg).get_result()

  @parameterized.expand([
      ({'changes': 0},),
      ({'changes': [0]},),
      ({'changes': [{'author': 0}]},),
      ({'changes': [{'author': {}}]},),
      ({'changes': [{'author': {'email': 0}}]},),
      ({'changes': [{'author': {'email': ''}}]},),
      ({'changes': [{'author': {'email': 'a@example.com'}, 'repo_url': 0}]},),
      ({'swarming': []},),
      ({'swarming': {'junk': 1}},),
      ({'swarming': {'recipe': []}},),
  ])
  def test_validate_build_parameters(self, parameters):
    with self.assertRaises(errors.InvalidInputError):
      swarming.validate_build_parameters(parameters)

  def test_shared_cache(self):
    self.builder_cfg.caches.add(path='builder', name='shared_builder_cache')

    slices = self.prepare_task_def(test_util.build())['task_slices']
    self.assertEqual(
        slices[0]['properties']['caches'], [
            {'path': 'cache/a', 'name': 'a'},
            {'path': 'cache/builder', 'name': 'shared_builder_cache'},
            {'path': 'cache/git_cache', 'name': 'git_chromium'},
            {'path': 'cache/out', 'name': 'build_chromium'},
        ]
    )

  def test_dimensions_and_cache_fallback(self):
    # Creates 4 task_slices by modifying the buildercfg in 2 ways:
    # - Add two named caches, one expiring at 60 seconds, one at 360 seconds.
    # - Add an optional builder dimension, expiring at 120 seconds.
    #
    # This ensures the combination of these features works correctly, and that
    # multiple 'caches' dimensions can be injected.
    self.builder_cfg.caches.add(
        path='builder',
        name='shared_builder_cache',
        wait_for_warm_cache_secs=60,
    )
    self.builder_cfg.caches.add(
        path='second',
        name='second_cache',
        wait_for_warm_cache_secs=360,
    )
    self.builder_cfg.dimensions.append("120:opt:ional")

    slices = self.prepare_task_def(test_util.build())['task_slices']

    self.assertEqual(4, len(slices))
    for t in slices:
      # They all use the same cache definitions.
      self.assertEqual(
          t['properties']['caches'], [
              {'path': u'cache/a', 'name': u'a'},
              {'path': u'cache/builder', 'name': u'shared_builder_cache'},
              {'path': u'cache/git_cache', 'name': u'git_chromium'},
              {'path': u'cache/out', 'name': u'build_chromium'},
              {'path': u'cache/second', 'name': u'second_cache'},
          ]
      )

    # But the dimensions are different. 'opt' and 'caches' are injected.
    self.assertEqual(
        slices[0]['properties']['dimensions'], [
            {u'key': u'caches', u'value': u'second_cache'},
            {u'key': u'caches', u'value': u'shared_builder_cache'},
            {u'key': u'cores', u'value': u'8'},
            {u'key': u'opt', u'value': u'ional'},
            {u'key': u'os', u'value': u'Ubuntu'},
            {u'key': u'pool', u'value': u'Chrome'},
        ]
    )
    self.assertEqual(slices[0]['expiration_secs'], '60')

    # One 'caches' expired. 'opt' and one 'caches' are still injected.
    self.assertEqual(
        slices[1]['properties']['dimensions'], [
            {u'key': u'caches', u'value': u'second_cache'},
            {u'key': u'cores', u'value': u'8'},
            {u'key': u'opt', u'value': u'ional'},
            {u'key': u'os', u'value': u'Ubuntu'},
            {u'key': u'pool', u'value': u'Chrome'},
        ]
    )
    # 120-60
    self.assertEqual(slices[1]['expiration_secs'], '60')

    # 'opt' expired, one 'caches' remains.
    self.assertEqual(
        slices[2]['properties']['dimensions'], [
            {u'key': u'caches', u'value': u'second_cache'},
            {u'key': u'cores', u'value': u'8'},
            {u'key': u'os', u'value': u'Ubuntu'},
            {u'key': u'pool', u'value': u'Chrome'},
        ]
    )
    # 360-120
    self.assertEqual(slices[2]['expiration_secs'], '240')

    # The cold fallback; the last 'caches' expired.
    self.assertEqual(
        slices[3]['properties']['dimensions'], [
            {u'key': u'cores', u'value': u'8'},
            {u'key': u'os', u'value': u'Ubuntu'},
            {u'key': u'pool', u'value': u'Chrome'},
        ]
    )
    # 3600-360
    self.assertEqual(slices[3]['expiration_secs'], '3240')

  def test_cache_fallback_fail_multiple_task_slices(self):
    # Make it so the swarming task template has two task_slices, which is
    # incompatible with a builder using wait_for_warm_cache_secs.
    self.task_template['task_slices'].append(
        self.task_template['task_slices'][0]
    )
    self.builder_cfg.caches.add(
        path='builder',
        name='shared_builder_cache',
        wait_for_warm_cache_secs=60,
    )

    with self.assertRaises(errors.InvalidInputError):
      self.prepare_task_def(test_util.build())

  def test_execution_timeout(self):
    self.builder_cfg.execution_timeout_secs = 120

    slices = self.prepare_task_def(test_util.build())['task_slices']

    self.assertEqual(slices[0]['properties']['execution_timeout_secs'], '120')

  def test_expiration(self):
    self.builder_cfg.expiration_secs = 120

    slices = self.prepare_task_def(test_util.build())['task_slices']

    self.assertEqual(2, len(slices))
    self.assertEqual(slices[0]['expiration_secs'], '60')
    self.assertEqual(slices[1]['expiration_secs'], '60')

  def test_auto_builder_dimension(self):
    self.builder_cfg.auto_builder_dimension = project_config_pb2.YES

    slices = self.prepare_task_def(test_util.build())['task_slices']
    self.assertEqual(
        slices[0]['properties']['dimensions'], [
            {u'key': u'builder', u'value': u'linux'},
            {u'key': u'caches', u'value': linux_CACHE_NAME},
            {u'key': u'cores', u'value': u'8'},
            {u'key': u'os', u'value': u'Ubuntu'},
            {u'key': u'pool', u'value': u'Chrome'},
        ]
    )

  def test_auto_builder_dimension_already_set(self):
    self.builder_cfg.auto_builder_dimension = project_config_pb2.YES
    self.builder_cfg.dimensions.append('builder:custom')

    slices = self.prepare_task_def(test_util.build())['task_slices']
    self.assertEqual(
        slices[0]['properties']['dimensions'], [
            {u'key': u'builder', u'value': u'custom'},
            {u'key': u'caches', u'value': linux_CACHE_NAME},
            {u'key': u'cores', u'value': u'8'},
            {u'key': u'os', u'value': u'Ubuntu'},
            {u'key': u'pool', u'value': u'Chrome'},
        ]
    )

  def test_unset_dimension(self):
    self.builder_cfg.dimensions[:] = ['cores:']

    slices = self.prepare_task_def(test_util.build())['task_slices']
    dim_keys = {d['key'] for d in slices[0]['properties']['dimensions']}
    self.assertNotIn('cores', dim_keys)

  def test_overall(self):
    self.patch(
        'components.auth.get_current_identity',
        autospec=True,
        return_value=auth.Identity('user', 'john@example.com')
    )
    self.patch(
        'swarming._is_migrating_builder_prod_async',
        autospec=True,
        return_value=future(True)
    )

    build = test_util.build(
        for_creation=True,
        id=1,
        number=1,
        builder=build_pb2.BuilderID(
            project='chromium', bucket='try', builder='linux'
        ),
        input=dict(
            properties=bbutil.dict_to_struct({'a': 'b'}),
            gerrit_changes=[
                dict(
                    host='chromium-review.googlesource.com',
                    project='chromium/src',
                    change=1234,
                    patchset=5,
                ),
            ],
        ),
    )

    build.parameters['changes'] = [{
        'author': {'email': 'bob@example.com'},
        'repo_url': 'https://chromium.googlesource.com/chromium/src',
    }]

    actual = self.prepare_task_def(build)

    expected_input_properties = {
        'a': 'b',
        'blamelist': ['bob@example.com'],
        'buildbucket': {
            'hostname': 'cr-buildbucket.appspot.com',
            'build': {
                'project':
                    'chromium',
                'bucket':
                    'luci.chromium.try',
                'created_by':
                    'anonymous:anonymous',
                'created_ts':
                    1448841600000000,
                'id':
                    '1',
                'tags': [
                    'build_address:luci.chromium.try/linux/1',
                    'builder:linux',
                    'buildset:1',
                ],
            },
        },
        'buildername': 'linux',
        'buildnumber': 1,
        'predefined-property': 'x',
        'predefined-property-bool': True,
        'repository': 'https://chromium.googlesource.com/chromium/src',
        '$recipe_engine/buildbucket': {
            'hostname': 'cr-buildbucket.appspot.com',
            'build': {
                'id': '1',
                'builder': {
                    'project': 'chromium',
                    'bucket': 'try',
                    'builder': 'linux',
                },
                'number': 1,
                'tags': [{'value': '1', 'key': 'buildset'}],
                'input': {
                    'gerritChanges': [{
                        'host': 'chromium-review.googlesource.com',
                        'project': 'chromium/src',
                        'change': '1234',
                        'patchset': '5',
                    }],
                },
                'infra': {
                    'buildbucket': {'serviceConfigRevision': 'template_rev'},
                    'logdog': {
                        'hostname': 'logdog.example.com',
                        'project': 'chromium',
                        'prefix': 'bb',
                    },
                    'recipe': {
                        'name': 'recipe',
                        'cipdPackage': 'infra/recipe_bundle',
                    },
                    'swarming': {
                        'hostname': 'swarming.example.com',
                        'taskId': 'deadbeef',
                        'taskServiceAccount': 'robot@example.com'
                    },
                },
                'createdBy': 'anonymous:anonymous',
                'createTime': '2015-11-30T00:00:00Z',
            },
        },
        '$recipe_engine/runtime': {
            'is_experimental': False,
            'is_luci': True,
        },
    }
    expected_swarming_props_def = {
        'env': [{
            'key': 'BUILDBUCKET_EXPERIMENTAL',
            'value': 'FALSE',
        }],
        'execution_timeout_secs':
            '3600',
        'extra_args': [
            'cook',
            '-recipe',
            'recipe',
            '-properties',
            api_common.properties_to_json(expected_input_properties),
            '-logdog-project',
            'chromium',
        ],
        'dimensions': [
            {'key': 'cores', 'value': '8'},
            {'key': 'os', 'value': 'Ubuntu'},
            {'key': 'pool', 'value': 'Chrome'},
        ],
        'caches': [
            {'path': 'cache/a', 'name': 'a'},
            {'path': 'cache/builder', 'name': linux_CACHE_NAME},
            {'path': 'cache/git_cache', 'name': 'git_chromium'},
            {'path': 'cache/out', 'name': 'build_chromium'},
        ],
        'cipd_input': {
            'packages': [
                {
                    'package_name': 'infra/test/bar/${os_ver}',
                    'path': '.',
                    'version': 'latest',
                },
                {
                    'package_name': 'infra/test/foo/${platform}',
                    'path': 'third_party',
                    'version': 'stable',
                },
                {
                    'package_name': 'infra/recipe_bundle',
                    'path': 'kitchen-checkout',
                    'version': 'refs/heads/master',
                },
            ],
        },
    }
    # The swarming template has fallback.
    props_def_first = copy.deepcopy(expected_swarming_props_def)
    props_def_first[u'dimensions'].append({
        u'key': u'caches',
        u'value': linux_CACHE_NAME,
    })
    props_def_first[u'dimensions'].sort(key=lambda x: (x[u'key'], x[u'value']))
    expected = {
        'name':
            'bb-1-chromium-linux',
        'priority':
            '108',
        'tags': [
            'build_address:luci.chromium.try/linux/1',
            'buildbucket_bucket:chromium/try',
            'buildbucket_build_id:1',
            'buildbucket_hostname:cr-buildbucket.appspot.com',
            'buildbucket_template_canary:0',
            'buildbucket_template_revision:template_rev',
            'builder:linux',
            'buildertag:yes',
            'buildset:1',
            'commontag:yes',
            (
                'log_location:logdog://luci-logdog-dev.appspot.com/chromium/'
                'buildbucket/cr-buildbucket.appspot.com/1/+/annotations'
            ),
            'luci_project:chromium',
            'recipe_name:recipe',
            'recipe_package:infra/recipe_bundle',
        ],
        'pool_task_template':
            'CANARY_NEVER',
        'task_slices': [
            {
                'expiration_secs': '60',
                'properties': props_def_first,
                'wait_for_capacity': False,
            },
            {
                'expiration_secs': '3540',
                'properties': expected_swarming_props_def,
                'wait_for_capacity': False,
            },
        ],
        'pubsub_topic':
            'projects/testbed-test/topics/swarming',
        'pubsub_userdata':
            json.dumps({
                'created_ts': 1448841600000000,
                'swarming_hostname': 'swarming.example.com',
                'build_id': 1L,
            },
                       sort_keys=True),
        'service_account':
            'robot@example.com',
        'numerical_value_for_coverage_in_format_obj':
            42,
    }
    self.assertEqual(test_util.ununicode(actual), expected)

    self.assertEqual(build.url, 'https://milo.example.com/b/1')

    self.assertEqual(
        build.proto.infra.swarming.task_service_account, 'robot@example.com'
    )

    self.assertEqual(
        build.proto.infra.buildbucket.service_config_revision, 'template_rev'
    )

    self.assertEqual(
        build.proto.infra.logdog.hostname, 'luci-logdog-dev.appspot.com'
    )
    self.assertEqual(build.proto.infra.logdog.project, 'chromium')
    self.assertEqual(
        build.proto.infra.logdog.prefix,
        'buildbucket/cr-buildbucket.appspot.com/1'
    )
    self.assertEqual(build.proto.input.properties['predefined-property'], 'x')
    self.assertNotIn('buildbucket', build.proto.input.properties)
    self.assertNotIn('$recipe_engine/buildbucket', build.proto.input.properties)
    self.assertEqual(build.proto.infra.recipe.name, 'recipe')

  def test_experimental(self):
    build = test_util.build(
        for_creation=True,
        input=dict(experimental=True),
    )
    actual = self.prepare_task_def(build)

    env = actual['task_slices'][0]['properties']['env']
    self.assertIn({
        'key': 'BUILDBUCKET_EXPERIMENTAL',
        'value': 'TRUE',
    }, env)

  def test_without_template(self):
    self.task_template = None
    self.task_template_canary = None

    with self.assertRaises(swarming.TemplateNotFound):
      self.prepare_task_def(test_util.build())

  def test_max_dimensions(self):
    self.builder_cfg.dimensions.extend([
        '60:opt1:ional',
        '120:opt2:ional',
        '240:opt3:ional',
        '300:opt4:ional',
        '360:opt5:ional',
        '420:opt6:ional',
    ])
    self.prepare_task_def(test_util.build())

  def test_bad_request_dimensions(self):
    self.builder_cfg.dimensions.extend([
        '60:opt1:ional',
        '120:opt2:ional',
        '240:opt3:ional',
        '300:opt4:ional',
        '360:opt5:ional',
        '420:opt6:ional',
        '480:opt7:ional',
    ])
    with self.assertRaises(errors.InvalidInputError):
      self.prepare_task_def(test_util.build())

  def test_canary_template(self):
    build = test_util.build(id=1, for_creation=True)
    build.canary_preference = model.CanaryPreference.CANARY

    actual = self.prepare_task_def(build)
    self.assertTrue(actual['name'].endswith('-canary'))

  def test_no_canary_template_explicit(self):
    build = test_util.build()
    build.canary_preference = model.CanaryPreference.CANARY

    self.task_template_canary = None
    with self.assertRaises(errors.InvalidInputError):
      self.prepare_task_def(build)

  @mock.patch('swarming._should_use_canary_template', autospec=True)
  def test_no_canary_template_implicit(self, should_use_canary_template):
    should_use_canary_template.return_value = True
    self.task_template_canary = None
    self.builder_cfg.task_template_canary_percentage.value = 54

    build = test_util.build()
    build.canary_preference = model.CanaryPreference.AUTO
    self.prepare_task_def(build)

    self.assertFalse(build.proto.infra.buildbucket.canary)
    should_use_canary_template.assert_called_with(54)

  def test_override_cfg(self):
    build = test_util.build()
    build.parameters['swarming'] = {
        'override_builder_cfg': {
            # Override cores dimension.
            'dimensions': ['cores:16'],
            'recipe': {'name': 'bob'},
        },
    }

    self.prepare_task_def(build)

    slices = self.prepare_task_def(test_util.build())['task_slices']
    self.assertIn(
        {'key': 'cores', 'value': '16'},
        slices[0]['properties']['dimensions'],
    )

  def test_override_dimensions(self):
    build = test_util.build(
        for_creation=True,
        infra=dict(
            buildbucket=dict(
                requested_dimensions=[
                    dict(key='cores', value='16'),
                ]
            ),
        ),
    )

    self.prepare_task_def(build)

    slices = self.prepare_task_def(test_util.build())['task_slices']
    self.assertIn(
        {'key': 'cores', 'value': '16'},
        slices[0]['properties']['dimensions'],
    )

  def test_override_cfg_malformed(self):
    build = test_util.build()
    build.parameters['swarming'] = {'override_builder_cfg': []}
    with self.assertRaises(errors.InvalidInputError):
      self.prepare_task_def(build)

    build = test_util.build()
    build.parameters['swarming'] = {'override_builder_cfg': {'name': 'x'}}
    with self.assertRaises(errors.InvalidInputError):
      self.prepare_task_def(build)

    build = test_util.build()
    build.parameters['swarming'] = {'override_builder_cfg': {'mixins': ['x']}}
    with self.assertRaises(errors.InvalidInputError):
      self.prepare_task_def(build)

    build = test_util.build()
    build.parameters['swarming'] = {'override_builder_cfg': {'blabla': 'x'}}
    with self.assertRaises(errors.InvalidInputError):
      self.prepare_task_def(build)

    # Remove a required dimension.
    build = test_util.build()
    build.parameters['swarming'] = {
        'override_builder_cfg': {'dimensions': ['pool:']}
    }
    with self.assertRaises(errors.InvalidInputError):
      self.prepare_task_def(build)

    # Override build numbers
    build = test_util.build()
    build.parameters['swarming'] = {
        'override_builder_cfg': {'build_numbers': False}
    }
    with self.assertRaises(errors.InvalidInputError):
      self.prepare_task_def(build)

  def test_on_leased_build(self):
    build = test_util.build()
    build.lease_key = 12345
    with self.assertRaises(errors.InvalidInputError):
      self.prepare_task_def(build)

  def test_generate_build_url(self):
    build = test_util.build(id=1, for_creation=True)
    self.assertEqual(
        swarming._generate_build_url('milo.example.com', build),
        'https://milo.example.com/b/1',
    )

    self.assertEqual(
        swarming._generate_build_url(None, build),
        ('https://swarming.example.com/task?id=deadbeef')
    )


class CreateTaskTest(BaseTest):

  def setUp(self):
    super(CreateTaskTest, self).setUp()
    self.patch('components.net.json_request_async', autospec=True)
    self.patch('components.auth.delegate_async', return_value=future('blah'))

    self.build_token = 'beeff00d'
    self.patch(
        'tokens.generate_build_token', autospec=True, return_value='deadbeef'
    )

    self.task_def = {'is_task_def': True, 'task_slices': [{
        'properties': {},
    }]}

    self.build = test_util.build(id=1, created_by='user:john@example.com')
    self.build.swarming_task_key = None
    with self.build.mutate_infra() as infra:
      infra.swarming.task_id = ''
    self.build.put()

  def test_success(self):
    expected_task_def = self.task_def.copy()
    expected_secrets = launcher_pb2.BuildSecrets(build_token=self.build_token)
    expected_task_def[u'task_slices'][0][u'properties'][u'secret_bytes'] = (
        base64.b64encode(expected_secrets.SerializeToString())
    )

    net.json_request_async.return_value = future({'task_id': 'x'})
    swarming._create_swarming_task(1, self.task_def)

    actual_task_def = net.json_request_async.call_args[1]['payload']
    self.assertEqual(actual_task_def, expected_task_def)

    self.assertEqual(
        net.json_request_async.call_args[0][0],
        'https://swarming.example.com/_ah/api/swarming/v1/tasks/new'
    )

    # Test delegation token params.
    self.assertEqual(
        auth.delegate_async.mock_calls, [
            mock.call(
                services=[u'https://swarming.example.com'],
                audience=[auth.Identity('user', 'test@localhost')],
                impersonate=auth.Identity('user', 'john@example.com'),
                tags=['buildbucket:bucket:chromium/try'],
            )
        ]
    )

  def test_already_exists(self):
    with self.build.mutate_infra() as infra:
      infra.swarming.task_id = 'exists'
    self.build.put()

    swarming._create_swarming_task(1, self.task_def)
    self.assertFalse(net.json_request_async.called)

  @mock.patch('swarming.cancel_task', autospec=True)
  def test_already_exists_after_creation(self, cancel_task):

    @ndb.tasklet
    def json_request_async(*_args, **_kwargs):
      with self.build.mutate_infra() as infra:
        infra.swarming.task_id = 'deadbeef'
      yield self.build.put_async()
      raise ndb.Return({'task_id': 'new task'})

    net.json_request_async.side_effect = json_request_async

    swarming._create_swarming_task(1, self.task_def)
    cancel_task.assert_called_with('swarming.example.com', 'new task')

  def test_http_401(self):
    net.json_request_async.return_value = future_exception(
        net.AuthError('forbidden', 401, None)
    )

    swarming._create_swarming_task(1, self.task_def)

    build = self.build.key.get()
    self.assertEqual(build.status, common_pb2.INFRA_FAILURE)
    self.assertEqual(
        build.proto.summary_markdown,
        r'Swarming responded with HTTP 401: `forbidden`'
    )

  def test_http_500(self):
    net.json_request_async.return_value = future_exception(
        net.Error('internal', 500, None)
    )

    with self.assertRaises(net.Error):
      swarming._create_swarming_task(1, self.task_def)


class CancelTest(BaseTest):

  def setUp(self):
    super(CancelTest, self).setUp()

    self.json_response = None

    def json_request_async(*_, **__):
      if self.json_response is not None:
        return future(self.json_response)
      self.fail('unexpected outbound request')  # pragma: no cover

    self.patch(
        'components.net.json_request_async',
        autospec=True,
        side_effect=json_request_async
    )

  def test_cancel_task(self):
    self.json_response = {'ok': True}
    swarming.cancel_task('swarming.example.com', 'deadbeef')
    net.json_request_async.assert_called_with(
        (
            'https://swarming.example.com/'
            '_ah/api/swarming/v1/task/deadbeef/cancel'
        ),
        method='POST',
        scopes=net.EMAIL_SCOPE,
        delegation_token=None,
        payload={'kill_running': True},
        deadline=None,
        max_attempts=None,
    )

  def test_cancel_running_task(self):
    self.json_response = {
        'was_running': True,
        'ok': False,
    }
    swarming.cancel_task('swarming.example.com', 'deadbeef')


class SyncBuildTest(BaseTest):

  @parameterized.expand([
      ({
          'task_result': None,
          'status': common_pb2.INFRA_FAILURE,
          'end_time': test_util.dt2ts(NOW),
      },),
      ({
          'task_result': {'state': 'PENDING'},
          'status': common_pb2.SCHEDULED,
      },),
      ({
          'task_result': {
              'state': 'RUNNING',
              'started_ts': '2018-01-29T21:15:02.649750',
          },
          'status': common_pb2.STARTED,
          'start_time': tspb(seconds=1517260502, nanos=649750000),
      },),
      ({
          'task_result': {
              'state': 'COMPLETED',
              'started_ts': '2018-01-29T21:15:02.649750',
              'completed_ts': '2018-01-30T00:15:18.162860',
          },
          'build_run_result': {},
          'status': common_pb2.SUCCESS,
          'start_time': tspb(seconds=1517260502, nanos=649750000),
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state':
                  'COMPLETED',
              'started_ts':
                  '2018-01-29T21:15:02.649750',
              'completed_ts':
                  '2018-01-30T00:15:18.162860',
              'bot_dimensions': [
                  {'key': 'os', 'value': ['Ubuntu', 'Trusty']},
                  {'key': 'pool', 'value': ['luci.chromium.try']},
                  {'key': 'id', 'value': ['bot1']},
              ],
          },
          'build_run_result': {
              'annotationUrl':
                  'logdog://logdog.example.com/chromium/prefix/+/annotations',
              'annotations': {},
          },
          'status':
              common_pb2.SUCCESS,
          'bot_dimensions': [
              common_pb2.StringPair(key='id', value='bot1'),
              common_pb2.StringPair(key='os', value='Trusty'),
              common_pb2.StringPair(key='os', value='Ubuntu'),
              common_pb2.StringPair(key='pool', value='luci.chromium.try'),
          ],
          'start_time':
              tspb(seconds=1517260502, nanos=649750000),
          'end_time':
              tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state': 'COMPLETED',
              'failure': True,
              'started_ts': '2018-01-29T21:15:02.649750',
              'completed_ts': '2018-01-30T00:15:18.162860',
          },
          'build_run_result': {},
          'status': common_pb2.FAILURE,
          'start_time': tspb(seconds=1517260502, nanos=649750000),
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state': 'COMPLETED',
              'failure': True,
              'started_ts': '2018-01-29T21:15:02.649750',
              'completed_ts': '2018-01-30T00:15:18.162860',
          },
          'build_run_result': {
              'infraFailure': {
                  'type': 'BOOTSTRAPPER_ERROR',
                  'text': 'it is not good',
              },
          },
          'status': common_pb2.INFRA_FAILURE,
          'start_time': tspb(seconds=1517260502, nanos=649750000),
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state': 'COMPLETED',
              'failure': True,
              'started_ts': '2018-01-29T21:15:02.649750',
              'completed_ts': '2018-01-30T00:15:18.162860',
          },
          'build_run_result': None,
          'status': common_pb2.INFRA_FAILURE,
          'start_time': tspb(seconds=1517260502, nanos=649750000),
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state': 'COMPLETED',
              'failure': True,
              'started_ts': '2018-01-29T21:15:02.649750',
              'completed_ts': '2018-01-30T00:15:18.162860',
          },
          'build_run_result_error': swarming._BUILD_RUN_RESULT_CORRUPTED,
          'status': common_pb2.INFRA_FAILURE,
          'start_time': tspb(seconds=1517260502, nanos=649750000),
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state': 'COMPLETED',
              'failure': True,
              'internal_failure': True,
              'started_ts': '2018-01-29T21:15:02.649750',
              'completed_ts': '2018-01-30T00:15:18.162860',
          },
          'status': common_pb2.INFRA_FAILURE,
          'start_time': tspb(seconds=1517260502, nanos=649750000),
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state': 'BOT_DIED',
              'started_ts': '2018-01-29T21:15:02.649750',
              'abandoned_ts': '2018-01-30T00:15:18.162860',
          },
          'status': common_pb2.INFRA_FAILURE,
          'start_time': tspb(seconds=1517260502, nanos=649750000),
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state': 'TIMED_OUT',
              'started_ts': '2018-01-29T21:15:02.649750',
              'completed_ts': '2018-01-30T00:15:18.162860',
          },
          'status': common_pb2.INFRA_FAILURE,
          'is_timeout': True,
          'start_time': tspb(seconds=1517260502, nanos=649750000),
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state': 'EXPIRED',
              'abandoned_ts': '2018-01-30T00:15:18.162860',
          },
          'status': common_pb2.INFRA_FAILURE,
          'is_resource_exhaustion': True,
          'is_timeout': True,
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state': 'KILLED',
              'abandoned_ts': '2018-01-30T00:15:18.162860',
          },
          'status': common_pb2.CANCELED,
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state': 'CANCELED',
              'abandoned_ts': '2018-01-30T00:15:18.162860',
          },
          'status': common_pb2.CANCELED,
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      ({
          'task_result': {
              'state': 'NO_RESOURCE',
              'abandoned_ts': '2018-01-30T00:15:18.162860',
          },
          'status': common_pb2.INFRA_FAILURE,
          'is_resource_exhaustion': True,
          'end_time': tspb(seconds=1517271318, nanos=162860000),
      },),
      # NO_RESOURCE with abandoned_ts before creation time.
      (
          {
              'task_result': {
                  'state': 'NO_RESOURCE',
                  'abandoned_ts': '2015-11-29T00:15:18.162860',
              },
              'status': common_pb2.INFRA_FAILURE,
              'is_resource_exhaustion': True,
              'end_time': test_util.dt2ts(NOW),
          },
      ),
  ])
  def test_sync(self, case):
    logging.info('test case: %s', case)
    self.patch(
        'swarming._load_build_run_result_async',
        autospec=True,
        return_value=future((
            case.get('build_run_result'),
            case.get('build_run_result_error', None),
        )),
    )
    build = test_util.build(id=1)
    build.put()

    swarming._sync_build_async(
        1, case['task_result'], 'luci.chromium.ci', 'linux-rel'
    ).get_result()

    build = build.key.get()
    bp = build.proto
    self.assertEqual(bp.status, case['status'])
    self.assertEqual(
        bp.status_details.HasField('timeout'),
        case.get('is_timeout', False),
    )
    self.assertEqual(
        bp.status_details.HasField('resource_exhaustion'),
        case.get('is_resource_exhaustion', False)
    )

    self.assertEqual(bp.start_time, case.get('start_time', tspb(0)))
    self.assertEqual(bp.end_time, case.get('end_time', tspb(0)))
    self.assertEqual(
        list(build.parse_infra().swarming.bot_dimensions),
        case.get('bot_dimensions', [])
    )


class LoadBuildRunResultTest(BaseTest):

  @mock.patch('swarming.isolate.fetch_async')
  def test_success(self, fetch_isolate_async):
    self.assertEqual(
        swarming._load_build_run_result_async({}, 'luci.chromium.ci',
                                              'linux-rel').get_result(),
        (None, None),
    )

    expected = {
        'infra_failure': {'text': 'not good'},
    }
    fetch_isolate_async.side_effect = [
        # isolated
        future(
            json.dumps({
                'files': {
                    swarming._BUILD_RUN_RESULT_FILENAME: {
                        'h': 'deadbeef',
                        's': 2048,
                    },
                },
            })
        ),
        # build-run-result.json
        future(json.dumps(expected)),
    ]
    actual, error = swarming._load_build_run_result_async({
        'id': 'taskid',
        'outputs_ref': {
            'isolatedserver': 'https://isolate.example.com',
            'namespace': 'default-gzip',
            'isolated': 'badcoffee',
        },
    }, 'luci.chromium.ci', 'linux-rel').get_result()
    self.assertIsNone(error)
    self.assertEqual(expected, actual)
    fetch_isolate_async.assert_any_call(
        isolate.Location(
            'isolate.example.com',
            'default-gzip',
            'badcoffee',
        )
    )
    fetch_isolate_async.assert_any_call(
        isolate.Location(
            'isolate.example.com',
            'default-gzip',
            'deadbeef',
        )
    )

  @mock.patch('swarming.isolate.fetch_async')
  def test_too_large(self, fetch_isolate_async):
    fetch_isolate_async.return_value = future(
        json.dumps({
            'files': {
                swarming._BUILD_RUN_RESULT_FILENAME: {
                    'h': 'deadbeef',
                    's': 1 + (2 << 20),
                },
            },
        })
    )
    actual, error = swarming._load_build_run_result_async({
        'id': 'taskid',
        'outputs_ref': {
            'isolatedserver': 'https://isolate.example.com',
            'namespace': 'default-gzip',
            'isolated': 'badcoffee',
        },
    }, 'luci.chromium.ci', 'linux-rel').get_result()
    self.assertEqual(error, swarming._BUILD_RUN_RESULT_TOO_LARGE)
    self.assertIsNone(actual)

  @mock.patch('swarming.isolate.fetch_async')
  def test_no_result(self, fetch_isolate_async):
    # isolated only, without the result
    fetch_isolate_async.return_value = future(
        json.dumps({
            'files': {'soemthing_else.txt': {
                'h': 'deadbeef',
                's': 2048,
            }},
        })
    )
    actual, error = swarming._load_build_run_result_async({
        'id': 'taskid',
        'outputs_ref': {
            'isolatedserver': 'https://isolate.example.com',
            'namespace': 'default-gzip',
            'isolated': 'badcoffee',
        },
    }, 'luci.chromium.ci', 'linux-rel').get_result()
    self.assertIsNone(error)
    self.assertIsNone(actual)

  def test_non_https_server(self):
    run_result, error = swarming._load_build_run_result_async({
        'id': 'taskid',
        'outputs_ref': {
            'isolatedserver': 'http://isolate.example.com',
            'namespace': 'default-gzip',
            'isolated': 'badcoffee',
        },
    }, 'luci.chromium.ci', 'linux-rel').get_result()
    self.assertEqual(error, swarming._BUILD_RUN_RESULT_CORRUPTED)
    self.assertIsNone(run_result)

  @mock.patch('swarming.isolate.fetch_async')
  def test_load_build_run_result_invalid_json(self, fetch_isolate_async):
    fetch_isolate_async.return_value = future('{"incomplete_json')
    run_result, error = swarming._load_build_run_result_async({
        'id': 'taskid',
        'outputs_ref': {
            'isolatedserver': 'https://isolate.example.com',
            'namespace': 'default-gzip',
            'isolated': 'badcoffee',
        },
    }, 'luci.chromium.ci', 'linux-rel').get_result()
    self.assertEqual(error, swarming._BUILD_RUN_RESULT_CORRUPTED)
    self.assertIsNone(run_result)

  @mock.patch('swarming.isolate.fetch_async')
  def test_isolate_error(self, fetch_isolate_async):
    fetch_isolate_async.side_effect = isolate.Error()
    with self.assertRaises(isolate.Error):
      swarming._load_build_run_result_async({
          'id': 'taskid',
          'outputs_ref': {
              'isolatedserver': 'https://isolate.example.com',
              'namespace': 'default-gzip',
              'isolated': 'badcoffee',
          },
      }, 'luci.chromium.ci', 'linux-rel').get_result()

  @mock.patch('swarming.isolate.fetch_async')
  def test_not_found(self, fetch_isolate_async):
    fetch_isolate_async.return_value = future(None)
    run_result, error = swarming._load_build_run_result_async({
        'id': 'taskid',
        'outputs_ref': {
            'isolatedserver': 'https://isolate.example.com',
            'namespace': 'default-gzip',
            'isolated': 'badcoffee',
        },
    }, 'luci.chromium.ci', 'linux-rel').get_result()
    self.assertEqual(error, swarming._BUILD_RUN_RESULT_CORRUPTED)
    self.assertIsNone(run_result)


class MigrationTest(BaseTest):

  def setUp(self):
    super(MigrationTest, self).setUp()

    self.json_response = None
    self.net_err_response = None

    def json_request_async(*_, **__):
      if self.net_err_response is not None:
        return future_exception(self.net_err_response)
      if self.json_response is not None:
        return future(self.json_response)
      self.fail('unexpected outbound request')  # pragma: no cover

    self.patch(
        'components.net.json_request_async',
        autospec=True,
        side_effect=json_request_async
    )

  def is_migrating_builder_prod(self, build):
    fut = swarming._is_migrating_builder_prod_async(self.builder_cfg, build)
    return fut.get_result()

  def test_no_master_name(self):
    self.builder_cfg.luci_migration_host = 'migration.example.com'
    self.assertIsNone(self.is_migrating_builder_prod(test_util.build()))
    self.assertFalse(net.json_request_async.called)

  def test_no_host(self):
    build = test_util.build(
        for_creation=True,
        input=dict(
            properties=bbutil.dict_to_struct({
                'mastername': 'tryserver.chromium.linux',
            })
        ),
    )
    self.assertIsNone(self.is_migrating_builder_prod(build))
    self.assertFalse(net.json_request_async.called)

  def test_prod(self):
    self.builder_cfg.luci_migration_host = 'migration.example.com'
    build = test_util.build(
        for_creation=True,
        input=dict(
            properties=bbutil.dict_to_struct({
                'mastername': 'tryserver.chromium.linux',
            })
        ),
    )
    self.json_response = {'luci_is_prod': True, 'bucket': 'luci.chromium.try'}
    self.assertTrue(self.is_migrating_builder_prod(build))
    self.assertTrue(net.json_request_async.called)

  def test_mastername_in_builder(self):
    self.builder_cfg.luci_migration_host = 'migration.example.com'
    self.builder_cfg.recipe.properties_j.append(
        'mastername:"tryserver.chromium.linux"'
    )
    self.json_response = {'luci_is_prod': True, 'bucket': 'luci.chromium.try'}
    self.assertTrue(self.is_migrating_builder_prod(test_util.build()))
    self.assertTrue(net.json_request_async.called)

  def test_custom_name(self):
    self.builder_cfg.luci_migration_host = 'migration.example.com'
    build = test_util.build(
        for_creation=True,
        input=dict(
            properties=bbutil.dict_to_struct({
                'luci_migration_master_name': 'custom_master',
                'mastername': 'ordinary_mastername',
            })
        ),
    )
    self.json_response = {'luci_is_prod': True, 'bucket': 'luci.chromium.try'}
    self.assertTrue(self.is_migrating_builder_prod(build))
    self.assertTrue(net.json_request_async.called)
    self.assertIn('custom_master', net.json_request_async.call_args[0][0])

  def test_404(self):
    self.builder_cfg.luci_migration_host = 'migration.example.com'
    build = test_util.build(
        for_creation=True,
        input=dict(
            properties=bbutil.dict_to_struct({
                'mastername': 'tryserver.chromium.linux',
            })
        ),
    )

    self.net_err_response = net.NotFoundError('nope', 404, 'can\'t find it')
    self.assertIsNone(self.is_migrating_builder_prod(build))

  def test_500(self):
    self.builder_cfg.luci_migration_host = 'migration.example.com'
    build = test_util.build(
        for_creation=True,
        input=dict(
            properties=bbutil.dict_to_struct({
                'mastername': 'tryserver.chromium.linux',
            })
        ),
    )

    self.net_err_response = net.Error('BOOM', 500, 'IT\'S BAD')
    self.assertIsNone(self.is_migrating_builder_prod(build))

  def test_no_is_prod_in_response(self):
    self.builder_cfg.luci_migration_host = 'migration.example.com'
    build = test_util.build(
        for_creation=True,
        input=dict(
            properties=bbutil.dict_to_struct({
                'mastername': 'tryserver.chromium.linux',
            })
        ),
    )

    self.net_err_response = None
    self.json_response = {'foo': True}
    self.assertIsNone(self.is_migrating_builder_prod(build))


class SubNotifyTest(BaseTest):

  def setUp(self):
    super(SubNotifyTest, self).setUp()
    self.handler = swarming.SubNotify(response=webapp2.Response())

    self.patch(
        'swarming._load_build_run_result_async',
        autospec=True,
        return_value=future(({}, False)),
    )

  def test_unpack_msg(self):
    self.assertEqual(
        self.handler.unpack_msg({
            'messageId':
                '1', 'data':
                    b64json({
                        'task_id':
                            'deadbeef',
                        'userdata':
                            json.dumps({
                                'created_ts': 1448841600000000,
                                'swarming_hostname': 'swarming.example.com',
                                'build_id': 1L,
                            })
                    })
        }), (
            'swarming.example.com', datetime.datetime(2015, 11, 30), 'deadbeef',
            1
        )
    )

  def test_unpack_msg_with_err(self):
    with self.assert_bad_message():
      self.handler.unpack_msg({})
    with self.assert_bad_message():
      self.handler.unpack_msg({'data': b64json([])})

    bad_data = [
        # Bad task id.
        {
            'userdata':
                json.dumps({
                    'created_ts': 1448841600000,
                    'swarming_hostname': 'swarming.example.com',
                })
        },

        # Bad swarming hostname.
        {
            'task_id': 'deadbeef',
        },
        {
            'task_id': 'deadbeef',
            'userdata': '{}',
        },
        {
            'task_id': 'deadbeef',
            'userdata': json.dumps({
                'swarming_hostname': 1,
            }),
        },

        # Bad creation time
        {
            'task_id':
                'deadbeef',
            'userdata':
                json.dumps({
                    'swarming_hostname': 'swarming.example.com',
                }),
        },
        {
            'task_id':
                'deadbeef',
            'userdata':
                json.dumps({
                    'created_ts': 'foo',
                    'swarming_hostname': 'swarming.example.com',
                }),
        },
    ]

    for data in bad_data:
      with self.assert_bad_message():
        self.handler.unpack_msg({'data': b64json(data)})

  def mock_request(self, user_data, task_id='deadbeef'):
    msg_data = b64json({
        'task_id': task_id,
        'userdata': json.dumps(user_data),
    })
    self.handler.request = mock.Mock(
        json={
            'message': {
                'messageId': '1',
                'data': msg_data,
            },
        }
    )

  @mock.patch('swarming._load_task_result_async', autospec=True)
  def test_post(self, load_task_result_async):
    build = test_util.build(id=1)
    build.put()

    self.mock_request({
        'build_id': 1L,
        'created_ts': 1448841600000000,
        'swarming_hostname': 'swarming.example.com',
    })

    load_task_result_async.return_value = future({
        'task_id': 'deadbeef',
        'state': 'COMPLETED',
    })

    self.handler.post()

    build = build.key.get()
    self.assertEqual(build.proto.status, common_pb2.SUCCESS)

  def test_post_with_different_swarming_hostname(self):
    build = test_util.build(id=1)
    build.put()

    self.mock_request({
        'build_id': 1L,
        'created_ts': 1448841600000000,
        'swarming_hostname': 'different-chromium.example.com',
    })

    with self.assert_bad_message(expect_redelivery=False):
      self.handler.post()

  def test_post_with_different_task_id(self):
    build = test_util.build(id=1)
    build.put()

    self.mock_request(
        {
            'build_id': 1L,
            'created_ts': 1448841600000000,
            'swarming_hostname': 'swarming.example.com',
        },
        task_id='deadbeefffffffffff',
    )

    with self.assert_bad_message(expect_redelivery=False):
      self.handler.post()

  def test_post_too_soon(self):
    build = test_util.build(id=1)
    with build.mutate_infra() as infra:
      infra.swarming.task_id = ''
    build.put()

    self.mock_request({
        'build_id': 1L,
        'created_ts': 1448841600000000,
        'swarming_hostname': 'swarming.example.com',
    })

    with self.assert_bad_message(expect_redelivery=True):
      self.handler.post()

  def test_post_without_task_id(self):
    self.mock_request(
        {
            'build_id': 1L,
            'created_ts': 1448841600000000,
            'swarming_hostname': 'swarming.example.com',
        },
        task_id=None,
    )

    with self.assert_bad_message(expect_redelivery=False):
      self.handler.post()

  def test_post_without_build_id(self):
    self.mock_request({
        'created_ts': 1448841600000000,
        'swarming_hostname': 'swarming.example.com',
    })
    with self.assert_bad_message(expect_redelivery=False):
      self.handler.post()

  def test_post_without_build(self):
    self.mock_request({
        'created_ts': 1438841600000000,
        'swarming_hostname': 'swarming.example.com',
        'build_id': 1L,
    })

    with self.assert_bad_message(expect_redelivery=False):
      self.handler.post()

  @contextlib.contextmanager
  def assert_bad_message(self, expect_redelivery=False):
    self.handler.bad_message = False
    err = exc.HTTPClientError if expect_redelivery else exc.HTTPOk
    with self.assertRaises(err):
      yield
    self.assertTrue(self.handler.bad_message)

  @mock.patch('swarming.SubNotify._process_msg', autospec=True)
  def test_dedup_messages(self, _process_msg):
    self.handler.request = mock.Mock(
        json={'message': {
            'messageId': '1',
            'data': b64json({}),
        }}
    )

    self.handler.post()
    self.handler.post()

    self.assertEquals(_process_msg.call_count, 1)


class CronUpdateTest(BaseTest):

  def setUp(self):
    super(CronUpdateTest, self).setUp()
    self.now += datetime.timedelta(minutes=5)

    self.patch(
        'swarming._load_build_run_result_async',
        autospec=True,
        return_value=future(({}, False)),
    )

  @mock.patch('swarming._load_task_result_async', autospec=True)
  def test_sync_build_async(self, load_task_result_async):
    load_task_result_async.return_value = future({
        'state': 'RUNNING',
    })

    build = test_util.build()
    build.put()

    swarming.CronUpdateBuilds().update_build_async(build).get_result()
    build = build.key.get()
    self.assertEqual(build.proto.status, common_pb2.STARTED)
    self.assertFalse(build.proto.HasField('end_time'))

    load_task_result_async.return_value = future({
        'state': 'COMPLETED',
    })

    swarming.CronUpdateBuilds().update_build_async(build).get_result()
    build = build.key.get()
    self.assertEqual(build.proto.status, common_pb2.SUCCESS)

  @mock.patch('swarming._load_task_result_async', autospec=True)
  def test_sync_build_async_no_task(self, load_task_result_async):
    load_task_result_async.return_value = future(None)

    build = test_util.build()
    build.put()
    swarming.CronUpdateBuilds().update_build_async(build).get_result()
    build = build.key.get()
    self.assertEqual(build.proto.status, common_pb2.INFRA_FAILURE)
    self.assertTrue(build.proto.summary_markdown)

  def test_sync_build_async_non_swarming(self):
    build = test_util.build(status=common_pb2.SCHEDULED)
    with build.mutate_infra() as infra:
      infra.ClearField('swarming')
    build.put()

    swarming.CronUpdateBuilds().update_build_async(build).get_result()

    build = build.key.get()
    self.assertEqual(build.proto.status, common_pb2.SCHEDULED)


class TestApplyIfTags(BaseTest):

  def test_no_replacement(self):
    obj = {
        'tags': [
            'something',
            'other_thing',
        ],
        'some key': ['value', {'other_key': 100}],
        'some other key': {
            'a': 'b',
            'b': 'c',
        },
    }
    self.assertEqual(swarming._apply_if_tags(obj), obj)

  def test_drop_if_tag(self):
    obj = {
        'tags': [
            'something',
            'other_thing',
        ],
        'some key': ['value', {
            'other_key': 100,
            '#if-tag': 'something',
        }],
        'some other key': {
            'a': 'b',
            'b': 'c',
        },
    }
    self.assertEqual(
        swarming._apply_if_tags(obj), {
            'tags': [
                'something',
                'other_thing',
            ],
            'some key': ['value', {
                'other_key': 100,
            }],
            'some other key': {
                'a': 'b',
                'b': 'c',
            },
        }
    )

  def test_drop_clause(self):
    obj = {
        'tags': [
            'something',
            'other_thing',
        ],
        'some key': ['value', {
            'other_key': 100,
            '#if-tag': 'something',
        }],
        'some other key': {
            'a': 'b',
            'b': 'c',
            '#if-tag': 'nope',
        },
    }
    self.assertEqual(
        swarming._apply_if_tags(obj), {
            'tags': [
                'something',
                'other_thing',
            ],
            'some key': ['value', {
                'other_key': 100,
            }],
        }
    )

  def test_parse_ts_without_usecs(self):
    actual = swarming._parse_ts('2018-02-23T04:22:45')
    expected = datetime.datetime(2018, 2, 23, 4, 22, 45)
    self.assertEqual(actual, expected)


def b64json(data):
  return base64.b64encode(json.dumps(data))
