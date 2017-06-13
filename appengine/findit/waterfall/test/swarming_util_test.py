# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import collections
import json
import mock
import os
import urllib
import zlib

from google.appengine.api.urlfetch_errors import DeadlineExceededError
from google.appengine.api.urlfetch_errors import DownloadError
from google.appengine.api.urlfetch_errors import ConnectionClosedError

from common.waterfall import buildbucket_client
from gae_libs.http.http_client_appengine import HttpClientAppengine
from infra_api_clients import logdog_util
from libs.http.retry_http_client import RetryHttpClient
from model.wf_config import FinditConfig
from model.wf_step import WfStep
from model.wf_try_bot_cache import WfTryBot
from model.wf_try_bot_cache import WfTryBotCache
from waterfall import swarming_util
from waterfall import waterfall_config
from waterfall.swarming_task_request import SwarmingTaskRequest
from waterfall.test import wf_testcase


class MockBuild(object):
  def __init__(self, response):
    self.response = response


MOCK_BUILDS = [(None, MockBuild({'tags': [
              'swarming_tag:log_location:logdog://host/project/path']}))]

ALL_BOTS = [{'bot_id': 'bot%d' % b} for b in range(10)]
SOME_BOTS = [{'bot_id': 'bot%d' % b} for b in range(3)]
ONE_BOT = [{'bot_id': 'bot%d' % b} for b in range(1)]


class MockTryJob(object):
  is_swarmbucket_build = True
  dimensions = ['os:OS', 'cpu:CPU']
  revision = 'a1b2c3d4'


class SwarmingHttpClient(RetryHttpClient):

  def __init__(self):
    self.get_responses = dict()
    self.post_responses = dict()

  def _GetData(self, data_type, file_name=None):
    file_name_map = {
        'build': 'sample_swarming_build_tasks.json',
        'step': 'sample_swarming_build_step_tasks.json',
        'task': 'sample_swarming_task.json'
    }
    file_name = file_name_map.get(data_type, file_name)

    swarming_tasks_file = os.path.join(
        os.path.dirname(__file__), 'data', file_name)
    with open(swarming_tasks_file, 'r') as f:
      return f.read()

  def _SetResponseForGetRequestIsolated(self, url, file_hash):
    self.get_responses[url] = self._GetData('isolated', file_hash)

  def _SetResponseForGetRequestSwarmingList(
      self, master_name, builder_name, build_number, step_name=None):
    if builder_name == 'download_failed':
      return

    url = ('https://%s/_ah/api/swarming/v1/tasks/'
           'list?tags=%s&tags=%s&tags=%s') % (
               FinditConfig().Get().swarming_settings.get('server_host'),
               urllib.quote('master:%s' % master_name),
               urllib.quote('buildername:%s' % builder_name),
               urllib.quote('buildnumber:%d' % build_number))

    if step_name:
      url += '&tags=%s' % urllib.quote('stepname:%s' % step_name)
      response = self._GetData('step')
    else:
      response = self._GetData('build')

    cursor_swarming_data = {
        'cursor': None,
        'items': [],
        'state': 'all',
        'limit': 100,
        'sort': 'created_ts'
    }
    cursor_url = ('%s&cursor=thisisacursor') % url

    self.get_responses[url] = response
    self.get_responses[cursor_url] = json.dumps(cursor_swarming_data)

  def _SetResponseForGetRequestSwarmingResult(self, task_id):
    url = ('https://%s/_ah/api/swarming/v1/task/%s/result') % (
        FinditConfig().Get().swarming_settings.get('server_host'), task_id)

    response = self._GetData('task')
    self.get_responses[url] = response

  def _SetResponseForPostRequest(self, isolated_hash):
    if isolated_hash == 'not found':
      response = '{"content":"eJyrrgUAAXUA+Q=="}'
    else:
      response = self._GetData('isolated', isolated_hash)

    self.post_responses[isolated_hash] = response

  def _Get(self, url, *_):
    if url in self.get_responses:
      return 200, self.get_responses[url]
    return 404, 'Download Failed!'

  def _Post(self, url, data, *_):
    data = json.loads(data)
    if data and data.get('digest') and data['digest'] in self.post_responses:
      return 200, self.post_responses[data['digest']]
    return 404, 'Download Failed!'

  def _Put(self, *_):  # pragma: no cover
    pass


class _LoggedHttpClient(RetryHttpClient):

  def __init__(self):
    self.responses = collections.defaultdict(dict)
    self.requests = {}

  def _Get(self, url, _, headers):
    self.requests[url] = ('get', None, headers)
    return self.responses.get('get', {}).get(url)

  def _Post(self, url, responses, _, headers):
    self.requests[url] = ('post', responses, headers)
    return self.responses.get('post', {}).get(url, (None, 404))

  def _Put(self, *_):  # pragma: no cover
    pass

  def SetResponse(self, method, url, content=None, status_code=200):
    self.responses[method][url] = (status_code, content)

  def GetRequest(self, url):
    return self.requests.get(url)


class SwarmingUtilTest(wf_testcase.WaterfallTestCase):

  def setUp(self):
    super(SwarmingUtilTest, self).setUp()
    self.http_client = SwarmingHttpClient()
    self.logged_http_client = _LoggedHttpClient()
    self.buildbucket_id = '88123'
    self.step_name = 'browser_tests on platform'

  def testGetSwarmingTaskRequest(self):
    task_request_json = {
        'expiration_secs': 2,
        'name': 'name',
        'parent_task_id': 'pti',
        'priority': 1,
        'properties': {
            'command': 'cmd',
            'dimensions': [{'key': 'd', 'value': 'dv'}],
            'env': [{'key': 'e', 'value': 'ev'}],
            'execution_timeout_secs': 4,
            'extra_args': ['--flag'],
            'grace_period_secs': 5,
            'idempotent': True,
            'inputs_ref': {
                'isolated': 'i',
                'isolatedserver': 'is',
                'namespace': 'ns',
            },
            'io_timeout_secs': 3,
        },
        'tags': ['tag'],
        'user': 'user',
        'pubsub_topic': None,
        'pubsub_auth_token': None,
        'pubsub_userdata': None,
    }
    task_id = '1'
    url = ('https://chromium-swarm.appspot.com/'
           '_ah/api/swarming/v1/task/%s/request' % task_id)
    self.logged_http_client.SetResponse(
        'get', url, json.dumps(task_request_json), 200)

    task_request = swarming_util.GetSwarmingTaskRequest(
        task_id, self.logged_http_client)

    self.assertEqual(task_request_json, task_request.Serialize())

  @mock.patch.object(swarming_util, '_SendRequestToServer',
                     return_value=(None, {'code': 1, 'message': 'error'}))
  def testGetSwarmingTaskRequestError(self, _):
    self.assertIsNone(
        swarming_util.GetSwarmingTaskRequest('task_id1', HttpClientAppengine()))

  def testTriggerSwarmingTask(self):
    request = SwarmingTaskRequest()
    request.expiration_secs = 2
    request.name = 'name'
    request.parent_task_id = 'pti'
    request.priority = 1
    request.tags = ['tag']
    request.user = 'user'
    request.command = 'cmd'
    request.dimensions = [{'key': 'd', 'value': 'dv'}]
    request.env = [{'key': 'e', 'value': 'ev'}]
    request.execution_timeout_secs = 4
    request.extra_args = ['--flag']
    request.grace_period_secs = 5
    request.idempotent = True
    request.inputs_ref = {'isolated': 'i'}
    request.io_timeout_secs = 3

    url = 'https://chromium-swarm.appspot.com/_ah/api/swarming/v1/tasks/new'
    self.logged_http_client.SetResponse(
        'post', url, json.dumps({'task_id': '1'}), 200)

    expected_task_request_json = {
        'expiration_secs': 72000,
        'name': 'name',
        'parent_task_id': 'pti',
        'priority': 150,
        'properties': {
            'command': 'cmd',
            'dimensions': [{'key': 'd', 'value': 'dv'}],
            'env': [{'key': 'e', 'value': 'ev'}],
            'execution_timeout_secs': 4,
            'extra_args': ['--flag'],
            'grace_period_secs': 5,
            'idempotent': True,
            'inputs_ref': {'isolated': 'i'},
            'io_timeout_secs': 3,
        },
        'tags': ['tag', 'findit:1', 'project:Chromium', 'purpose:post-commit'],
        'user': 'user',
        'pubsub_topic': None,
        'pubsub_auth_token': None,
        'pubsub_userdata': None,
    }

    task_id, error = swarming_util.TriggerSwarmingTask(
        request, self.logged_http_client)
    self.assertEqual('1', task_id)
    self.assertIsNone(error)

    method, data, _ = self.logged_http_client.GetRequest(url)
    self.assertEqual('post', method)
    self.assertEqual(expected_task_request_json, json.loads(data))

  @mock.patch.object(swarming_util, '_SendRequestToServer',
                     return_value=(None, {'code': 1, 'message': 'error'}))
  def testTriggerSwarmingTaskError(self, _):
    request = SwarmingTaskRequest()
    task_id, error = swarming_util.TriggerSwarmingTask(
        request, HttpClientAppengine())
    self.assertIsNone(task_id)
    self.assertIsNotNone(error)

  def testGetIsolatedDataForFailedBuild(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 223
    failed_steps = {
        'a_tests': {
            'current_failure': 2,
            'first_failure': 0
        },
        'unit_tests': {
            'current_failure': 2,
            'first_failure': 0
        },
        'compile': {
            'current_failure': 2,
            'first_failure': 0
        }
    }

    self.http_client._SetResponseForGetRequestSwarmingList(
        master_name, builder_name, build_number)
    result = swarming_util.GetIsolatedDataForFailedBuild(
        master_name, builder_name, build_number, failed_steps,
        self.http_client)

    expected_failed_steps = {
        'a_tests': {
            'current_failure': 2,
            'first_failure': 0,
            'list_isolated_data': [
                {
                    'digest': 'isolatedhashatests',
                    'namespace': 'default-gzip',
                    'isolatedserver': (
                        waterfall_config.GetSwarmingSettings().get(
                            'isolated_server'))
                }
            ]
        },
        'unit_tests': {
            'current_failure': 2,
            'first_failure': 0,
            'list_isolated_data': [
                {
                    'digest': 'isolatedhashunittests1',
                    'namespace': 'default-gzip',
                    'isolatedserver': (
                        waterfall_config.GetSwarmingSettings().get(
                            'isolated_server'))
                }
            ]
        },
        'compile': {
            'current_failure': 2,
            'first_failure': 0
        }
    }

    for step_name in failed_steps:
      step = WfStep.Get(master_name, builder_name, build_number, step_name)
      if step_name == 'compile':
        self.assertIsNone(step)
      else:
        self.assertIsNotNone(step)

    self.assertTrue(result)
    self.assertEqual(expected_failed_steps, failed_steps)

  def testGetIsolatedDataForFailedBuildDownloadFailed(self):
    master_name = 'm'
    builder_name = 'download_failed'
    build_number = 223
    failed_steps = {
        'a_tests': {
            'current_failure': 2,
            'first_failure': 0
        },
        'unit_tests': {
            'current_failure': 2,
            'first_failure': 0
        }
    }

    self.http_client._SetResponseForGetRequestSwarmingList(
        master_name, builder_name, build_number)

    result = swarming_util.GetIsolatedDataForFailedBuild(
        master_name, builder_name, build_number, failed_steps,
        self.http_client)
    expected_failed_steps = {
        'a_tests': {
            'current_failure': 2,
            'first_failure': 0
        },
        'unit_tests': {
            'current_failure': 2,
            'first_failure': 0
        }
    }
    self.assertFalse(result)
    self.assertEqual(expected_failed_steps, failed_steps)

  def testGetIsolatedDataForStep(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 223
    step_name = 'unit_tests'

    self.http_client._SetResponseForGetRequestSwarmingList(
        master_name, builder_name, build_number, step_name)
    data = swarming_util.GetIsolatedDataForStep(
        master_name, builder_name, build_number, step_name,
        self.http_client)
    expected_data = [
        {
            'digest': 'isolatedhashunittests',
            'namespace': 'default-gzip',
            'isolatedserver': waterfall_config.GetSwarmingSettings().get(
                'isolated_server')
        }
    ]
    self.assertEqual(expected_data, data)

  def testGetIsolatedDataForStepNotOnlyFailure(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 223
    step_name = 'unit_tests'

    self.http_client._SetResponseForGetRequestSwarmingList(
        master_name, builder_name, build_number, step_name)
    data = swarming_util.GetIsolatedDataForStep(
        master_name, builder_name, build_number, step_name,
        self.http_client, only_failure=False)
    expected_data = [
        {
            'digest': 'isolatedhashunittests',
            'namespace': 'default-gzip',
            'isolatedserver': waterfall_config.GetSwarmingSettings().get(
                'isolated_server')
        },
        {
            'digest': 'isolatedhashunittests1',
            'namespace': 'default-gzip',
            'isolatedserver': waterfall_config.GetSwarmingSettings().get(
                'isolated_server')
        }
    ]
    self.assertEqual(sorted(expected_data), sorted(data))

  def testGetIsolatedDataForStepFailed(self):
    master_name = 'm'
    builder_name = 'download_failed'
    build_number = 223
    step_name = 's1'

    self.http_client._SetResponseForGetRequestSwarmingList(
        master_name, builder_name, build_number, step_name)

    task_ids = swarming_util.GetIsolatedDataForStep(
        master_name, builder_name, build_number, step_name,
        self.http_client)
    expected_task_ids = []

    self.assertEqual(expected_task_ids, task_ids)

  def testDownloadTestResults(self):
    isolated_data = {
        'digest': 'shard1_isolated',
        'namespace': 'default-gzip',
        'isolatedserver': waterfall_config.GetSwarmingSettings().get(
            'isolated_server')
    }
    isolated_storage_url = waterfall_config.GetSwarmingSettings().get(
        'isolated_storage_url')
    self.http_client._SetResponseForPostRequest('shard1_isolated')
    self.http_client._SetResponseForPostRequest('shard1_url')
    self.http_client._SetResponseForGetRequestIsolated(
        'https://%s/default-gzip/shard1' % isolated_storage_url, 'shard1')

    result, error = swarming_util._DownloadTestResults(
        isolated_data, self.http_client)

    expected_result = json.loads(zlib.decompress(
        self.http_client._GetData('isolated', 'shard1')))
    self.assertEqual(expected_result, result)
    self.assertIsNone(error)

  def testDownloadTestResultsFailedForSecondHash(self):
    isolated_data = {
        'digest': 'not found',
        'namespace': 'default-gzip',
        'isolatedserver': waterfall_config.GetSwarmingSettings().get(
            'isolated_server')
    }

    result, error = swarming_util._DownloadTestResults(
        isolated_data, self.http_client)

    self.assertIsNone(result)
    self.assertIsNotNone(error)

  def testDownloadTestResultsFailedForParsingSecondHash(self):
    isolated_data = {
        'digest': 'not found',
        'namespace': 'default-gzip',
        'isolatedserver': waterfall_config.GetSwarmingSettings().get(
            'isolated_server')
    }

    self.http_client._SetResponseForPostRequest('not found')
    result, error = swarming_util._DownloadTestResults(
        isolated_data, self.http_client)

    self.assertIsNone(result)
    self.assertIsNone(error)

  def testDownloadTestResultsFailedForFileUrl(self):
    isolated_data = {
        'digest': 'shard1_isolated',
        'namespace': 'default-gzip',
        'isolatedserver': waterfall_config.GetSwarmingSettings().get(
            'isolated_server')
    }
    self.http_client._SetResponseForPostRequest('shard1_isolated')
    result, error = swarming_util._DownloadTestResults(
        isolated_data, self.http_client)

    self.assertIsNone(result)
    self.assertIsNotNone(error)

  def testDownloadTestResultsFailedForFile(self):
    isolated_data = {
        'digest': 'shard1_isolated',
        'namespace': 'default-gzip',
        'isolatedserver': waterfall_config.GetSwarmingSettings().get(
            'isolated_server')
    }
    self.http_client._SetResponseForPostRequest('shard1_isolated')
    self.http_client._SetResponseForPostRequest('shard1_url')
    result, error = swarming_util._DownloadTestResults(
        isolated_data, self.http_client)

    self.assertIsNone(result)
    self.assertIsNone(error)

  def testRetrieveShardedTestResultsFromIsolatedServer(self):
    isolated_data = [
        {
            'digest': 'shard1_isolated',
            'namespace': 'default-gzip',
            'isolatedserver': waterfall_config.GetSwarmingSettings().get(
                'isolated_server')
        },
        {
            'digest': 'shard2_isolated',
            'namespace': 'default-gzip',
            'isolatedserver': waterfall_config.GetSwarmingSettings().get(
                'isolated_server')
        },
        {
            'digest': 'shard3_isolated',
            'namespace': 'default-gzip',
            'isolatedserver': waterfall_config.GetSwarmingSettings().get(
                'isolated_server')
        }
    ]
    isolated_storage_url = waterfall_config.GetSwarmingSettings().get(
        'isolated_storage_url')
    self.http_client._SetResponseForPostRequest('shard1_isolated')
    self.http_client._SetResponseForPostRequest('shard1_url')
    self.http_client._SetResponseForGetRequestIsolated(
        'https://%s/default-gzip/shard1' % isolated_storage_url, 'shard1')
    self.http_client._SetResponseForPostRequest('shard2_isolated')
    self.http_client._SetResponseForPostRequest('shard2_url')
    self.http_client._SetResponseForGetRequestIsolated(
        'https://%s/default-gzip/shard2' % isolated_storage_url, 'shard2')
    self.http_client._SetResponseForPostRequest('shard3_isolated')
    self.http_client._SetResponseForPostRequest('shard3_url')
    self.http_client._SetResponseForGetRequestIsolated(
        'https://%s/default-gzip/shard3' % isolated_storage_url, 'shard3')

    result = swarming_util.RetrieveShardedTestResultsFromIsolatedServer(
        isolated_data, self.http_client)
    expected_results_file = os.path.join(
        os.path.dirname(__file__), 'data', 'expected_collect_results')
    with open(expected_results_file, 'r') as f:
      expected_result = json.loads(f.read())

    self.assertEqual(expected_result, result)

  def testRetrieveShardedTestResultsFromIsolatedServerSingleShard(self):
    isolated_data = [
        {
            'digest': 'shard1_isolated',
            'namespace': 'default-gzip',
            'isolatedserver': waterfall_config.GetSwarmingSettings().get(
                'isolated_server')
        }
    ]
    self.http_client._SetResponseForPostRequest('shard1_isolated')
    self.http_client._SetResponseForPostRequest('shard1_url')
    self.http_client._SetResponseForGetRequestIsolated(
        'https://%s/default-gzip/shard1' %
        waterfall_config.GetSwarmingSettings().get('isolated_storage_url'),
        'shard1')

    result = swarming_util.RetrieveShardedTestResultsFromIsolatedServer(
        isolated_data, self.http_client)

    expected_result = json.loads(zlib.decompress(
        self.http_client._GetData('isolated', 'shard1')))
    self.assertEqual(expected_result, result)

  def testRetrieveShardedTestResultsFromIsolatedServerFailed(self):
    isolated_data = [
        {
            'digest': 'shard1_isolated',
            'namespace': 'default-gzip',
            'isolatedserver': waterfall_config.GetSwarmingSettings().get(
                'isolated_server')
        }
    ]

    result = swarming_util.RetrieveShardedTestResultsFromIsolatedServer(
        isolated_data, self.http_client)

    self.assertIsNone(result)

  def testGetSwarmingTaskResultById(self):
    task_id = '2944afa502297110'

    self.http_client._SetResponseForGetRequestSwarmingResult(task_id)

    data, error = swarming_util.GetSwarmingTaskResultById(
        task_id, self.http_client)

    expected_outputs_ref = {
        'isolatedserver': waterfall_config.GetSwarmingSettings().get(
            'isolated_server'),
        'namespace': 'default-gzip',
        'isolated': 'shard1_isolated'
    }

    self.assertEqual('COMPLETED', data['state'])
    self.assertEqual(expected_outputs_ref, data['outputs_ref'])
    self.assertIsNone(error)

  @mock.patch.object(swarming_util, '_SendRequestToServer',
                     return_value=(None, {'code': 1, 'message': 'error'}))
  def testGetSwarmingTaskResultByIdError(self, _):
    data, error = swarming_util.GetSwarmingTaskResultById(
        'task_id', HttpClientAppengine())
    self.assertEqual({}, data)
    self.assertIsNotNone(error)

  def testGetSwarmingTaskFailureLog(self):
    outputs_ref = {
        'isolatedserver': waterfall_config.GetSwarmingSettings().get(
            'isolated_server'),
        'namespace': 'default-gzip',
        'isolated': 'shard1_isolated'
    }

    self.http_client._SetResponseForPostRequest('shard1_isolated')
    self.http_client._SetResponseForPostRequest('shard1_url')
    self.http_client._SetResponseForGetRequestIsolated(
        'https://%s/default-gzip/shard1' % (
            waterfall_config.GetSwarmingSettings().get('isolated_storage_url')),
        'shard1')

    result, error = swarming_util.GetSwarmingTaskFailureLog(
        outputs_ref, self.http_client)

    expected_result = json.loads(zlib.decompress(
        self.http_client._GetData('isolated', 'shard1')))
    self.assertEqual(expected_result, result)
    self.assertIsNone(error)

  def testRetrieveOutputJsonFileGetDirectly(self):
    output_json_content = ('{"content": "eJyrVkpLzMwpLUotVrKKVgpJLS4xV'
                           'IrVUVAqS8zJTFGyUigpKk2tBQDr9wxZ"}')

    failure_log = swarming_util._RetrieveOutputJsonFile(
        output_json_content, self.http_client)

    expected_failure_log = {
        'failures': ['Test1'],
        'valid': True
    }

    self.assertEqual(expected_failure_log, failure_log)

  def testGetTagValueInvalidTag(self):
    tags = ['a:1', 'b:2']
    self.assertIsNone(swarming_util.GetTagValue(tags, 'c'))

  def testGenerateIsolatedDataOutputsrefNone(self):
    self.assertEqual({}, swarming_util._GenerateIsolatedData(None))

  def testFetchOutputJsonInfoFromIsolatedServerReturnNone(self):
    self.assertIsNone(swarming_util._FetchOutputJsonInfoFromIsolatedServer(
        None, self.http_client))

  @mock.patch.object(
      RetryHttpClient, 'Get', side_effect=ConnectionClosedError())
  def testSendRequestToServerConnectionClosedError(self, _):
    content, error = swarming_util._SendRequestToServer(
        'http://www.someurl.url', HttpClientAppengine())
    self.assertIsNone(content)
    self.assertEqual(
        error['code'], swarming_util.URLFETCH_CONNECTION_CLOSED_ERROR)

  @mock.patch.object(
      RetryHttpClient, 'Get', side_effect=DeadlineExceededError())
  def testSendRequestToServerDeadlineExceededError(self, _):
    content, error = swarming_util._SendRequestToServer(
        'http://www.someurl.com', HttpClientAppengine())
    self.assertIsNone(content)
    self.assertEqual(
        error['code'], swarming_util.URLFETCH_DEADLINE_EXCEEDED_ERROR)

  @mock.patch.object(RetryHttpClient, 'Get', side_effect=DownloadError())
  def testSendRequestToServerDownloadError(self, _):
    content, error = swarming_util._SendRequestToServer(
        'http://www.someurl.com', HttpClientAppengine())
    self.assertIsNone(content)
    self.assertEqual(error['code'], swarming_util.URLFETCH_DOWNLOAD_ERROR)

  def testGetBackoffSeconds(self):
    self.assertEqual(1, swarming_util._GetBackoffSeconds(1, 1, 1))
    self.assertEqual(2, swarming_util._GetBackoffSeconds(1, 2, 100))
    self.assertEqual(100, swarming_util._GetBackoffSeconds(1, 8, 100))

  @mock.patch.object(
      RetryHttpClient, 'Get', side_effect=ConnectionClosedError())
  def testSendRequestToServerRetryTimeout(self, _):
    override_swarming_settings = {
        'should_retry_server': True,
        'server_retry_timeout_hours': -1
    }
    self.UpdateUnitTestConfigSettings(
        'swarming_settings', override_swarming_settings)
    content, error = swarming_util._SendRequestToServer(
        'http://www.someurl.com', HttpClientAppengine())
    self.assertIsNone(content)
    self.assertEqual(
        error['code'], swarming_util.URLFETCH_CONNECTION_CLOSED_ERROR)
    self.assertTrue(error['retry_timeout'])

  @mock.patch.object(swarming_util, 'GetSwarmingTaskResultById',
                     return_value=({'outputs_ref': 'ref'}, None))
  @mock.patch.object(swarming_util, 'GetSwarmingTaskFailureLog',
                     return_value=(None, 'error'))
  def testGetIsolatedOutputForTaskIsolatedError(self, *_):
    self.assertIsNone(swarming_util.GetIsolatedOutputForTask(None, None))

  @mock.patch.object(swarming_util, 'GetSwarmingTaskResultById',
                     return_value=({'a': []}, None))
  def testGetIsolatedOutputForTaskNoOutputRef(self, _):
    self.assertIsNone(swarming_util.GetIsolatedOutputForTask(None, None))

  @mock.patch.object(swarming_util, 'GetSwarmingTaskResultById',
                     return_value=(None, 'error'))
  def testGetIsolatedOutputForTaskDataError(self, _):
    self.assertIsNone(swarming_util.GetIsolatedOutputForTask(None, None))

  def testGetIsolatedOutputForTask(self):
    task_id = '2944afa502297110'
    self.http_client._SetResponseForGetRequestSwarmingResult(task_id)
    self.http_client._SetResponseForPostRequest('shard1_isolated')
    self.http_client._SetResponseForPostRequest('shard1_url')
    self.http_client._SetResponseForGetRequestIsolated(
        'https://%s/default-gzip/shard1' % (
            waterfall_config.GetSwarmingSettings().get('isolated_storage_url')),
        'shard1')

    result = swarming_util.GetIsolatedOutputForTask(task_id, self.http_client)

    expected_result = json.loads(zlib.decompress(
        self.http_client._GetData('isolated', 'shard1')))
    self.assertEqual(expected_result, result)

  def testGetSwarmingBotCountsNodimentsions(self):
    self.assertEqual({}, swarming_util.GetSwarmingBotCounts(None, None))

  @mock.patch.object(swarming_util, '_SendRequestToServer',
                     return_value=(None, {'code': 1, 'message': 'error'}))
  def testGetSwarmingBotCountsError(self, _):

    dimensions = {
      'os': 'OS',
      'cpu': 'cpu'
    }
    self.assertEqual({}, swarming_util.GetSwarmingBotCounts(
        dimensions, self.http_client))

  @mock.patch.object(swarming_util, '_SendRequestToServer')
  def testGetSwarmingBotCounts(self, mock_fn):

    dimensions = {
      'os': 'OS',
      'cpu': 'cpu'
    }

    content_data = {
        'count': '10',
        'dead': '1',
        'quarantined': '0',
        'busy': '5'
    }
    mock_fn.return_value = (json.dumps(content_data), None)

    expected_counts = {
        'count': 10,
        'dead': 1,
        'quarantined': 0,
        'busy': 5,
        'available': 4
    }

    self.assertEqual(expected_counts, swarming_util.GetSwarmingBotCounts(
        dimensions, self.http_client))

  @mock.patch.object(
      buildbucket_client, 'GetTryJobs', return_value=MOCK_BUILDS)
  @mock.patch.object(logdog_util, '_GetLogLocationFromBuildbucketBuild',
                     return_value=('host', 'project', 'path'))
  @mock.patch.object(logdog_util, '_GetAnnotationsProtoForPath',
                     return_value='step')
  @mock.patch.object(logdog_util, '_GetStreamForStep',
                     return_value='log_stream')
  @mock.patch.object(logdog_util, 'GetStepLogForBuild',
                     return_value=json.dumps(wf_testcase.SAMPLE_STEP_METADATA))
  def testGetStepMetadata(self, *_):
    step_metadata = swarming_util.GetStepLog(
      self.buildbucket_id, self.step_name, self.http_client, 'step_metadata')
    self.assertEqual(step_metadata, wf_testcase.SAMPLE_STEP_METADATA)

  @mock.patch.object(
      buildbucket_client, 'GetTryJobs', return_value=MOCK_BUILDS)
  @mock.patch.object(logdog_util, '_GetAnnotationsProtoForPath',
                     return_value=None)
  def testGetStepMetadataStepNone(self, *_):
    step_metadata = swarming_util.GetStepLog(
      self.buildbucket_id, self.step_name, self.http_client, 'step_metadata')
    self.assertIsNone(step_metadata)

  @mock.patch.object(
      buildbucket_client, 'GetTryJobs', return_value=MOCK_BUILDS)
  @mock.patch.object(logdog_util, '_GetAnnotationsProtoForPath',
                     return_value='step')
  @mock.patch.object(logdog_util, '_GetLogLocationFromBuildbucketBuild',
                     return_value=('host', 'project','some/path'))
  @mock.patch.object(logdog_util, '_GetStreamForStep',
                     return_value=None)
  def testGetStepMetadataStreamNone(self, *_):
    step_metadata = swarming_util.GetStepLog(
      self.buildbucket_id, self.step_name, self.http_client, 'step_metadata')
    self.assertIsNone(step_metadata)

  @mock.patch.object(
      buildbucket_client, 'GetTryJobs', return_value=MOCK_BUILDS)
  @mock.patch.object(logdog_util, '_GetAnnotationsProtoForPath',
                     return_value='step')
  @mock.patch.object(logdog_util, '_GetLogLocationFromBuildbucketBuild',
                     return_value=('host', 'project','some/path'))
  @mock.patch.object(logdog_util, '_GetStreamForStep', return_value='stream')
  @mock.patch.object(logdog_util, 'GetStepLogForBuild',
                     return_value='log1/nlog2')
  def testGetStepLogStdio(self, *_):
    self.assertEqual('log1/nlog2', swarming_util.GetStepLog(
        self.buildbucket_id, self.step_name, self.http_client))

  @mock.patch.object(
      buildbucket_client, 'GetTryJobs', return_value=MOCK_BUILDS)
  @mock.patch.object(logdog_util, '_GetAnnotationsProtoForPath',
                     return_value=None)
  @mock.patch.object(logdog_util, '_GetLogLocationFromBuildbucketBuild',
                     return_value=('host', 'project','some/path'))
  def testGetStepLogStdioNoAnnotations(self, *_):
    self.assertIsNone(swarming_util.GetStepLog(
         self.buildbucket_id, self.step_name, self.http_client))

  @mock.patch.object(
      buildbucket_client, 'GetTryJobs', return_value=[(Exception(), None)])
  def testGetStepLogBuildbucketError(self, *_):
    self.assertIsNone(swarming_util.GetStepLog(
         self.buildbucket_id, self.step_name, self.http_client))

  def testUpdateAnalysisResult(self):
    analysis_result = {
        'failures': [
            {
                'step_name': 'another_step1',
                'flaky': True
            },
            {
                'tests': [
                  {
                      'last_pass': 123,
                      'first_failure': 123,
                      'suspected_cls': [],
                      'test_name': 'TestSuite1.test1'
                  },
                  {
                      'last_pass': 123,
                      'first_failure': 123,
                      'suspected_cls': [],
                      'test_name': 'TestSuite1.test2'
                  },
                  {
                      'last_pass': 123,
                      'first_failure': 123,
                      'suspected_cls': [],
                      'test_name': 'TestSuite1.test3'
                  },
                ],
                'step_name': self.step_name
            },
            {
                'step_name': 'another_step2'
            },
        ]
    }

    flaky_failures = {
        self.step_name: [
            'TestSuite1.test1', 'TestSuite1.test2']
    }

    all_flaked = swarming_util.UpdateAnalysisResult(
        analysis_result, flaky_failures)

    expected_result = {
        'failures': [
            {
                'step_name': 'another_step1',
                'flaky': True
            },
            {
                'tests': [
                    {
                        'last_pass': 123,
                        'first_failure': 123,
                        'suspected_cls': [],
                        'test_name': 'TestSuite1.test1',
                        'flaky': True
                    },
                    {
                        'last_pass': 123,
                        'first_failure': 123,
                        'suspected_cls': [],
                        'test_name': 'TestSuite1.test2',
                        'flaky': True
                    },
                    {
                        'last_pass': 123,
                        'first_failure': 123,
                        'suspected_cls': [],
                        'test_name': 'TestSuite1.test3'
                    },
                ],
                'step_name': self.step_name,
                'flaky': False
            },
            {
                'step_name': 'another_step2'
            },
        ]
    }

    self.assertFalse(all_flaked)
    self.assertEqual(expected_result, analysis_result)

  def testUpdateAnalysisResultAllFlaky(self):
    analysis_result = {
        'failures': [
            {
                'tests': [
                  {
                      'last_pass': 123,
                      'first_failure': 123,
                      'suspected_cls': [],
                      'test_name': 'TestSuite1.test1'
                  },
                  {
                      'last_pass': 123,
                      'first_failure': 123,
                      'suspected_cls': [],
                      'test_name': 'TestSuite1.test2'
                  }
                ],
                'step_name': self.step_name
            }
        ]
    }

    flaky_failures = {
        self.step_name: [
            'TestSuite1.test1', 'TestSuite1.test2']
    }

    all_flaked = swarming_util.UpdateAnalysisResult(
        analysis_result, flaky_failures)

    expected_result = {
        'failures': [
            {
                'tests': [
                    {
                        'last_pass': 123,
                        'first_failure': 123,
                        'suspected_cls': [],
                        'test_name': 'TestSuite1.test1',
                        'flaky': True
                    },
                    {
                        'last_pass': 123,
                        'first_failure': 123,
                        'suspected_cls': [],
                        'test_name': 'TestSuite1.test2',
                        'flaky': True
                    }
                ],
                'step_name': self.step_name,
                'flaky': True
            }
        ]
    }

    self.assertTrue(all_flaked)
    self.assertEqual(expected_result, analysis_result)

  def testUpdateAnalysisResultOnlyStep(self):
    analysis_result = {
        'failures': [
            {
                'step_name': 'another_step1'
            }
        ]
    }

    flaky_failures = {
        self.step_name: [
            'TestSuite1.test1', 'TestSuite1.test2']
    }

    all_flaked = swarming_util.UpdateAnalysisResult(
        analysis_result, flaky_failures)

    self.assertFalse(all_flaked)

  def testGetCacheName(self):
    cache_name_a = swarming_util.GetCacheName('luci.chromium.try',
                                              'LUCI linux_chromium_variable')
    cache_name_b = swarming_util.GetCacheName('luci.chromium.try',
                                              'LUCI win_chromium_variable')
    cache_name_c = swarming_util.GetCacheName('luci.chromium.ci',
                                              'LUCI win_chromium_variable')

    self.assertTrue(cache_name_a.startswith('builder_'))
    self.assertTrue(cache_name_b.startswith('builder_'))
    self.assertTrue(cache_name_c.startswith('builder_'))
    self.assertNotEqual(cache_name_a, cache_name_b)
    self.assertNotEqual(cache_name_a, cache_name_c)
    self.assertNotEqual(cache_name_b, cache_name_c)


  def testGetBot(self):
    class MockBuildbucketBuild(object):
      response = {
          'result_details_json': json.dumps({
              'swarming': {
                  'task_result':{
                      'bot_id': 'slave777-c4'
                  }
              }
          })
      }
    self.assertEqual('slave777-c4', swarming_util.GetBot(MockBuildbucketBuild))

  def testGetBotNotFound(self):
    class MockBuildbucketBuild(object):
      response = {
          'result_details_json': json.dumps({})
      }
    self.assertIsNone(swarming_util.GetBot(MockBuildbucketBuild))
    MockBuildbucketBuild.response = {}
    self.assertIsNone(swarming_util.GetBot(MockBuildbucketBuild))

  def testGetBuilderCacheName(self):
    class MockBuildbucketBuild(object):
      response = {
          'parameters_json': json.dumps({
              'swarming': {
                  'override_builder_cfg': {
                      'caches': [{
                          'path': 'builder',
                          'name':'builder_dummyhash'
                      }]
                  }
              }
          })
      }
    self.assertEqual('builder_dummyhash', swarming_util.GetBuilderCacheName(
        MockBuildbucketBuild))

  def testGetBuilderCacheNameNotFound(self):
    class MockBuildbucketBuild(object):
      response = {
          'parameters_json': json.dumps({
              'swarming': {}
          })
      }
    self.assertIsNone(swarming_util.GetBuilderCacheName(MockBuildbucketBuild))
    MockBuildbucketBuild.response = {
        'parameters_json': json.dumps({
            'swarming': {
                'override_builder_cfg': {
                    'caches': [{
                        'path': 'other_cache',
                        'name':'other_cache_name'
                    }]
                }
            }
        })
    }
    self.assertIsNone(swarming_util.GetBuilderCacheName(MockBuildbucketBuild))
    MockBuildbucketBuild.response = {}
    self.assertIsNone(swarming_util.GetBuilderCacheName(MockBuildbucketBuild))

  @mock.patch.object(swarming_util, '_SendRequestToServer')
  def testSelectWarmCacheNoOp(self, mock_fn):
    class MockTryJobBuildbot(object):
      is_swarmbucket_build = False
    try_job_buildbot = MockTryJobBuildbot()
    cache_name = 'some_other_cache_name'
    WfTryBotCache.Get(cache_name).recent_bots = ['slave1']
    swarming_util.AssignWarmCacheHost(try_job_buildbot, cache_name,
                                      SwarmingHttpClient())
    self.assertFalse(mock_fn.called)

  @mock.patch.object(swarming_util, '_SendRequestToServer',
                     return_value=(None, {'code': 1, 'message': 'error'}))
  def testGetAllBotsWithCacheError(self, _):
    dimensions = {
      'os': 'OS',
      'cpu': 'cpu'
    }
    self.assertEqual([], swarming_util.GetAllBotsWithCache(
        dimensions, 'cache_name', self.http_client))

  @mock.patch.object(swarming_util, '_SendRequestToServer')
  def testGetAllBotsWithCache(self, mock_fn):

    dimensions = {
      'os': 'OS',
      'cpu': 'cpu'
    }

    content_data = {
        'items': [
            {'bot_id': 'bot_1'}
        ]
    }
    mock_fn.return_value = (json.dumps(content_data), None)
    self.assertEqual(content_data['items'], swarming_util.GetAllBotsWithCache(
        dimensions, 'cache_name', self.http_client))

  def testOnlyAvailable(self):
    all_bots = [
        {'bot_id': 'bot1', 'task_id': '123abc000'},
        {'bot_id': 'bot2', 'is_dead': True},
        {'bot_id': 'bot3', 'quarantined': True},
        {'bot_id': 'bot4', 'deleted': True},
        {'bot_id': 'bot5'}
    ]
    self.assertEqual(
        [{'bot_id': 'bot5'}],
        swarming_util.OnlyAvailable(all_bots)
    )

  def testHaveCommitPositionInLocalGitCache(self):
    bots = [{'bot_id': 'bot%d' % i} for i in range(10)]
    bot5 = WfTryBot.Get('bot5')
    bot5.newest_synced_revision = 100
    bot5.put()
    self.assertEqual(
        [{'bot_id': 'bot5'}],
        swarming_util._HaveCommitPositionInLocalGitCache(bots, 1)
    )

  def testSortByDistanceToCommitPosition(self):
    cache_name = 'cache_name'
    cache_stats = WfTryBotCache.Get(cache_name)
    cache_stats.AddBot('bot1', 80, 80)
    cache_stats.AddBot('bot2', 90, 90)
    cache_stats.AddBot('bot3', 110, 110)
    cache_stats.AddBot('bot4', 120, 120)
    cache_stats.put()
    bots = [{'bot_id': 'bot%d' % i} for i in range(1, 5)]
    closest = swarming_util._ClosestEarlier(
        bots, cache_name, 70)
    self.assertFalse(closest)
    closest = swarming_util._ClosestLater(
        bots, cache_name, 70)
    self.assertEqual({'bot_id':'bot1'}, closest)

    sorted_bots = swarming_util._SortByDistanceToCommitPosition(
        bots, cache_name, 100, False)
    self.assertEqual({'bot_id':'bot2'}, sorted_bots[0])
    sorted_bots = swarming_util._SortByDistanceToCommitPosition(
        bots, cache_name, 121, False)
    self.assertEqual({'bot_id':'bot4'}, sorted_bots[0])


  @mock.patch('waterfall.swarming_util.GetAllBotsWithCache',
              return_value=ALL_BOTS)
  @mock.patch('waterfall.swarming_util.OnlyAvailable',
              return_value=SOME_BOTS)
  @mock.patch('waterfall.swarming_util._HaveCommitPositionInLocalGitCache',
              return_value=SOME_BOTS)
  @mock.patch('waterfall.swarming_util._ClosestEarlier',
              return_value=ONE_BOT[0])
  @mock.patch('waterfall.swarming_util._ClosestLater',
              return_value=ONE_BOT[0])
  @mock.patch('waterfall.swarming_util.CachedGitilesRepository.GetChangeLog')
  def testAssignWarmCacheHost(self, *_):
    cache_name = 'cache_name'
    tryjob = MockTryJob()
    swarming_util.AssignWarmCacheHost(tryjob, cache_name, self.http_client)

    # No bots with cache, check no bot id
    # No bots with rev, check bot_id of only bot with cache
    # No bots with earlier rev, check bot_id with earliest later rev
    # Bot with earlier rev gets assigned.

  @mock.patch('waterfall.swarming_util.GetAllBotsWithCache',
              return_value=ALL_BOTS)
  @mock.patch('waterfall.swarming_util.OnlyAvailable',
              return_value=SOME_BOTS)
  @mock.patch('waterfall.swarming_util._HaveCommitPositionInLocalGitCache',
              return_value=SOME_BOTS)
  @mock.patch('waterfall.swarming_util._ClosestEarlier',
              return_value=ONE_BOT[0])
  @mock.patch('waterfall.swarming_util.CachedGitilesRepository.GetChangeLog')
  def testAssignWarmCacheHostEarlier(self, *_):
    cache_name = 'cache_name'
    tryjob = MockTryJob()
    swarming_util.AssignWarmCacheHost(tryjob, cache_name, self.http_client)
    self.assertEqual('id:bot0', tryjob.dimensions[2])

  @mock.patch('waterfall.swarming_util.GetAllBotsWithCache',
              return_value=ALL_BOTS)
  @mock.patch('waterfall.swarming_util.OnlyAvailable',
              return_value=SOME_BOTS)
  @mock.patch('waterfall.swarming_util._HaveCommitPositionInLocalGitCache',
              return_value=SOME_BOTS)
  @mock.patch('waterfall.swarming_util._ClosestEarlier',
              return_value=None)
  @mock.patch('waterfall.swarming_util._ClosestLater',
              return_value=ONE_BOT[0])
  @mock.patch('waterfall.swarming_util.CachedGitilesRepository.GetChangeLog')
  def testAssignWarmCacheHostLater(self, *_):
    cache_name = 'cache_name'
    tryjob = MockTryJob()
    swarming_util.AssignWarmCacheHost(tryjob, cache_name, self.http_client)
    self.assertEqual('id:bot0', tryjob.dimensions[2])

  @mock.patch('waterfall.swarming_util.GetAllBotsWithCache',
              return_value=ALL_BOTS)
  @mock.patch('waterfall.swarming_util.OnlyAvailable',
              return_value=SOME_BOTS)
  @mock.patch('waterfall.swarming_util._HaveCommitPositionInLocalGitCache',
              return_value=ONE_BOT)
  @mock.patch('waterfall.swarming_util._ClosestEarlier',
              return_value=None)
  @mock.patch('waterfall.swarming_util._ClosestLater',
              return_value=None)
  @mock.patch('waterfall.swarming_util.CachedGitilesRepository.GetChangeLog')
  def testAssignWarmCacheHostNoCheckedOutRevision(self, *_):
    cache_name = 'cache_name'
    tryjob = MockTryJob()
    swarming_util.AssignWarmCacheHost(tryjob, cache_name, self.http_client)
    self.assertEqual('id:bot0', tryjob.dimensions[2])

  @mock.patch('waterfall.swarming_util.GetAllBotsWithCache',
              return_value=ALL_BOTS)
  @mock.patch('waterfall.swarming_util.OnlyAvailable',
              return_value=ONE_BOT)
  @mock.patch('waterfall.swarming_util._HaveCommitPositionInLocalGitCache',
              return_value=[])
  @mock.patch('waterfall.swarming_util.CachedGitilesRepository.GetChangeLog')
  def testAssignWarmCacheHostNoCachedRevision(self, *_):
    cache_name = 'cache_name'
    tryjob = MockTryJob()
    swarming_util.AssignWarmCacheHost(tryjob, cache_name, self.http_client)
    self.assertEqual('id:bot0', tryjob.dimensions[2])

  @mock.patch('waterfall.swarming_util.GetAllBotsWithCache',
              return_value=[])
  @mock.patch('waterfall.swarming_util.OnlyAvailable',
              return_value=[])
  @mock.patch('waterfall.swarming_util.CachedGitilesRepository.GetChangeLog')
  def disabled_testAssignWarmCacheHostBrandNewCache(
      self, *_):  # pragma: no cover.
    cache_name = 'cache_name'
    tryjob = MockTryJob()
    swarming_util.AssignWarmCacheHost(tryjob, cache_name, self.http_client)
    self.assertEqual(2, len(tryjob.dimensions))
