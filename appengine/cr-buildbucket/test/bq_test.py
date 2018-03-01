# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import datetime
import json
import os

from google.appengine.api import taskqueue
from google.appengine.ext import ndb

import webtest

from components import auth
from components import pubsub
from components import net
from testing_utils import testing
import mock

from test import test_util
import bq
import model
import v2

THIS_DIR = os.path.dirname(os.path.abspath(__file__))
APP_ROOT_DIR = os.path.dirname(THIS_DIR)


class BigQueryExportTest(testing.AppengineTestCase):
  taskqueue_stub_root_path = APP_ROOT_DIR

  def setUp(self):
    super(BigQueryExportTest, self).setUp()
    self.patch('components.net.json_request', autospec=True, return_value={})
    self.patch(
        'components.utils.utcnow', return_value=datetime.datetime(2018, 1, 1))


  def test_cron_export_builds_to_bq(self):
    builds = [
      mkbuild(
          id=1,
          status=model.BuildStatus.SCHEDULED),
      mkbuild(
          id=2,
          status=model.BuildStatus.STARTED,
          start_time=datetime.datetime(2018, 1, 1)),
      mkbuild(
          id=3,
          status=model.BuildStatus.COMPLETED,
          result=model.BuildResult.SUCCESS,
          complete_time=datetime.datetime(2018, 1, 1)),
      mkbuild(
          id=4,
          status=model.BuildStatus.COMPLETED,
          result=model.BuildResult.FAILURE,
          failure_reason=model.FailureReason.BUILD_FAILURE,
          complete_time=datetime.datetime(2018, 1, 1),
          experimental=True),
    ]
    ndb.put_multi(builds)
    taskqueue.Queue(bq._QUEUE_NAME).add([
      taskqueue.Task(
          method='PULL',
          tag=bq._task_tag(b.experimental),
          payload=json.dumps({'id': b.key.id()}),
      )
      for b in builds
    ])

    bq._process_pull_task_batch(False)
    net.json_request.assert_called_once_with(
        url=(
            'https://www.googleapis.com/bigquery/v2/'
            'projects/testbed-test/datasets/builds/tables/'
            'completed_beta/insertAll'
        ),
        method='POST',
        payload={
          'kind': 'bigquery#tableDataInsertAllRequest',
          'skipInvalidRows': False,
          'ignoreUnknownValues': False,
          'rows': [{
            'insertId': '3',
            'json': mock.ANY,
          }],
        },
        scopes=[
          'https://www.googleapis.com/auth/bigquery.insertdata',
          'https://www.googleapis.com/auth/bigquery',
          'https://www.googleapis.com/auth/cloud-platform',
        ],
        deadline=5 * 60,
    )

    rows = net.json_request.call_args_list[0][1]['payload']['rows']
    self.assertEqual(len(rows), 1)
    self.assertEqual(rows[0]['json']['id'], 3)
    self.assertEqual(rows[0]['json']['status'], 'SUCCESS')

    net.json_request.reset_mock()
    bq._process_pull_task_batch(True)
    net.json_request.assert_called_once_with(
        url=(
            'https://www.googleapis.com/bigquery/v2/'
            'projects/testbed-test/datasets/builds_experimental/tables/'
            'completed_beta/insertAll'
        ),
        method='POST',
        payload={
          'kind': 'bigquery#tableDataInsertAllRequest',
          'skipInvalidRows': False,
          'ignoreUnknownValues': False,
          'rows': [{
            'insertId': '4',
            'json': mock.ANY,
          }],
        },
        scopes=[
          'https://www.googleapis.com/auth/bigquery.insertdata',
          'https://www.googleapis.com/auth/bigquery',
          'https://www.googleapis.com/auth/cloud-platform',
        ],
        deadline=5 * 60,
    )

    rows = net.json_request.call_args_list[0][1]['payload']['rows']
    self.assertEqual(len(rows), 1)
    self.assertEqual(rows[0]['json']['id'], 4)
    self.assertEqual(rows[0]['json']['status'], 'FAILURE')

  def test_cron_export_builds_to_bq_unsupported(self):
    model.Build(
        id=1,
        bucket='foo',
        status=model.BuildStatus.COMPLETED,
        result=model.BuildResult.SUCCESS,
        create_time=datetime.datetime(2018, 1, 1),
        complete_time=datetime.datetime(2018, 1, 1),
    ).put()
    taskqueue.Queue(bq._QUEUE_NAME).add([taskqueue.Task(
        method='PULL',
        tag=bq._task_tag(False),
        payload=json.dumps({'id': 1}))
    ])
    bq._process_pull_task_batch(False)
    self.assertFalse(net.json_request.called)

  def test_cron_export_builds_to_bq_not_found(self):
    taskqueue.Queue(bq._QUEUE_NAME).add([taskqueue.Task(
        method='PULL',
        tag=bq._task_tag(False),
        payload=json.dumps({'id': 1}))
    ])
    bq._process_pull_task_batch(False)
    self.assertFalse(net.json_request.called)

  def test_cron_export_builds_to_bq_no_tasks(self):
    bq._process_pull_task_batch(False)
    self.assertFalse(net.json_request.called)

  @mock.patch(
      'google.appengine.api.taskqueue.Queue.delete_tasks', autospec=True)
  def test_cron_export_builds_to_bq_insert_errors(self, delete_tasks):
    builds = [
      mkbuild(
          id=i + 1,
          status=model.BuildStatus.COMPLETED,
          result=model.BuildResult.SUCCESS,
          complete_time=datetime.datetime(2018, 1, 1))
      for i in xrange(3)
    ]
    ndb.put_multi(builds)
    tasks = [
      taskqueue.Task(
          method='PULL',
          tag=bq._task_tag(False),
          payload=json.dumps({'id': b.key.id()}))
      for b in builds
    ]
    q = taskqueue.Queue(bq._QUEUE_NAME)
    q.add(tasks)

    net.json_request.return_value = {
      'insertErrors': [{
        'index': 1,
        'errors': [{'reason': 'bad', 'message': ':('}],
      }]
    }

    bq._process_pull_task_batch(False)
    self.assertTrue(net.json_request.called)

    # assert second task is not deleted
    deleted = delete_tasks.call_args[0][1]
    self.assertEqual(
      [t.payload for t in deleted],
      [tasks[0].payload, tasks[2].payload],
    )


def mkbuild(**kwargs):
  args = dict(
      id=1,
      project='chromium',
      bucket='luci.chromium.try',
      parameters={v2.BUILDER_PARAMETER: 'linux-rel'},
      created_by=auth.Identity('user', 'john@example.com'),
      create_time=datetime.datetime(2018, 1, 1),
  )
  args['parameters'].update(kwargs.pop('parameters', {}))
  args.update(kwargs)
  return model.Build(**args)
