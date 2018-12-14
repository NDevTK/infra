# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import datetime
import json

from components import utils
utils.fix_protobuf_package()

from google.appengine.ext import ndb
from google.protobuf import struct_pb2

import webtest

from components import pubsub
from testing_utils import testing

from test import test_util
import api_common
import main
import model
import notifications
import v2


class NotificationsTest(testing.AppengineTestCase):

  def setUp(self):
    super(NotificationsTest, self).setUp()

    self.app = webtest.TestApp(
        main.create_backend_app(), extra_environ={'REMOTE_ADDR': '127.0.0.1'}
    )

    self.patch(
        'notifications.enqueue_tasks_async',
        autospec=True,
        return_value=test_util.future(None)
    )
    self.patch(
        'bq.enqueue_pull_task_async',
        autospec=True,
        return_value=test_util.future(None)
    )

    self.patch(
        'google.appengine.api.app_identity.get_default_version_hostname',
        return_value='buildbucket.example.com',
        autospec=True
    )

    self.patch(
        'components.utils.utcnow', return_value=datetime.datetime(2017, 1, 1)
    )

    self.patch('components.pubsub.publish', autospec=True)

  def test_pubsub_callback(self):
    build = model.Build(
        id=1,
        bucket_id='chromium/try',
        create_time=datetime.datetime(2017, 1, 1),
        pubsub_callback=model.PubSubCallback(
            topic='projects/example/topics/buildbucket',
            user_data='hello',
            auth_token='secret',
        ),
        input_properties=struct_pb2.Struct(),
    )

    @ndb.transactional
    def txn():
      build.put()
      notifications.enqueue_notifications_async(build).get_result()

    txn()

    build = build.key.get()
    global_task_payload = {
        'id': 1,
        'mode': 'global',
    }
    callback_task_payload = {
        'id': 1,
        'mode': 'callback',
    }
    notifications.enqueue_tasks_async.assert_called_with(
        'backend-default', [
            {
                'url': '/internal/task/buildbucket/notify/1',
                'payload': json.dumps(global_task_payload, sort_keys=True),
                'age_limit_sec': model.BUILD_TIMEOUT.total_seconds(),
            },
            {
                'url': '/internal/task/buildbucket/notify/1',
                'payload': json.dumps(callback_task_payload, sort_keys=True),
                'age_limit_sec': model.BUILD_TIMEOUT.total_seconds(),
            },
        ]
    )

    self.app.post_json(
        '/internal/task/buildbucket/notify/1',
        params=global_task_payload,
        headers={'X-AppEngine-QueueName': 'backend-default'}
    )
    pubsub.publish.assert_called_with(
        'projects/testbed-test/topics/builds',
        json.dumps({
            'build': api_common.build_to_dict(build),
            'hostname': 'buildbucket.example.com',
        },
                   sort_keys=True),
        {'build_id': '1'},
    )

    self.app.post_json(
        '/internal/task/buildbucket/notify/1',
        params=callback_task_payload,
        headers={'X-AppEngine-QueueName': 'backend-default'}
    )
    pubsub.publish.assert_called_with(
        'projects/example/topics/buildbucket',
        json.dumps({
            'build': api_common.build_to_dict(build),
            'hostname': 'buildbucket.example.com',
            'user_data': 'hello',
        },
                   sort_keys=True),
        {
            'build_id': '1',
            'auth_token': 'secret',
        },
    )

  def test_no_pubsub_callback(self):
    build = model.Build(
        id=1,
        bucket_id='chromium/try',
        create_time=datetime.datetime(2017, 1, 1),
        input_properties=struct_pb2.Struct(),
    )

    @ndb.transactional
    def txn():
      build.put()
      notifications.enqueue_notifications_async(build).get_result()

    txn()

    build = build.key.get()
    global_task_payload = {
        'id': 1,
        'mode': 'global',
    }
    notifications.enqueue_tasks_async.assert_called_with(
        'backend-default', [
            {
                'url': '/internal/task/buildbucket/notify/1',
                'payload': json.dumps(global_task_payload, sort_keys=True),
                'age_limit_sec': model.BUILD_TIMEOUT.total_seconds(),
            },
        ]
    )

    self.app.post_json(
        '/internal/task/buildbucket/notify/1',
        params=global_task_payload,
        headers={'X-AppEngine-QueueName': 'backend-default'}
    )
    pubsub.publish.assert_called_with(
        'projects/testbed-test/topics/builds',
        json.dumps({
            'build': api_common.build_to_dict(build),
            'hostname': 'buildbucket.example.com',
        },
                   sort_keys=True),
        {'build_id': '1'},
    )
