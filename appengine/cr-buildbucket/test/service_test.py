# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import contextlib
import datetime

from components import auth
from components import utils
from google.appengine.ext import ndb
from google.appengine.ext.ndb import msgprop
from testing_utils import testing
import mock

from go.chromium.org.luci.buildbucket.proto import build_pb2
from go.chromium.org.luci.buildbucket.proto import common_pb2
from go.chromium.org.luci.buildbucket.proto import service_config_pb2
from test import test_util
from test.test_util import future
import config
import errors
import model
import notifications
import service
import swarming
import user


class BuildBucketServiceTest(testing.AppengineTestCase):

  TEST_BUCKETS = [
      'chromium/try',
      'chromium/luci',
      'chromium/master.foo',
      'chromium/master.bar',
  ]

  def setUp(self):
    super(BuildBucketServiceTest, self).setUp()

    self.current_identity = auth.Identity('service', 'unittest')
    self.patch(
        'components.auth.get_current_identity',
        side_effect=lambda: self.current_identity
    )
    self.now = datetime.datetime(2015, 1, 1)
    self.patch('components.utils.utcnow', side_effect=lambda: self.now)

    self.perms = test_util.mock_permissions(self)
    for b in self.TEST_BUCKETS:
      self.perms[b] = list(user.ALL_PERMISSIONS)

    test_util.put_empty_bucket('chromium', 'try')

    config.put_bucket(
        'chromium',
        'a' * 40,
        test_util.parse_bucket_cfg(
            '''
            name: "luci"
            swarming {
              builders {
                name: "linux"
                swarming_host: "chromium-swarm.appspot.com"
                build_numbers: YES
                recipe {
                  cipd_package: "infra/recipe_bundle"
                  cipd_version: "refs/heads/master"
                  name: "recipe"
                }
              }
            }
            '''
        ),
    )

    self.patch(
        'google.appengine.api.app_identity.get_default_version_hostname',
        autospec=True,
        return_value='buildbucket.example.com'
    )

    self.patch('tq.enqueue_async', autospec=True, return_value=future(None))
    self.patch(
        'config.get_settings_async',
        autospec=True,
        return_value=future(service_config_pb2.SettingsCfg())
    )
    self.patch(
        'swarming.cancel_task_transactionally_async',
        autospec=True,
        return_value=future(None)
    )

    self.patch('search.TagIndex.random_shard_index', return_value=0)

    test_util.build_bundle(id=1).infra.put()

  def mock_no_perm(self, perm):
    for perms in self.perms.values():
      perms.remove(perm)

  def put_many_builds(self, count=100, **build_proto_fields):
    builds = []
    build_ids = model.create_build_ids(utils.utcnow(), count)
    for build_id in build_ids:
      builds.append(self.classic_build(id=build_id, **build_proto_fields))
      self.now += datetime.timedelta(seconds=1)
    ndb.put_multi(builds)
    return builds

  @staticmethod
  def classic_build(**build_proto_fields):
    build = test_util.build(**build_proto_fields)
    build.is_luci = False
    return build

  #################################### GET #####################################

  def test_get(self):
    self.classic_build(id=1).put()
    build = service.get_async(1).get_result()
    self.assertEqual(build, build)

  def test_get_nonexistent_build(self):
    self.assertIsNone(service.get_async(42).get_result())

  def test_get_with_auth_error(self):
    self.mock_no_perm(user.PERM_BUILDS_GET)
    self.classic_build(id=1).put()
    with self.assertRaises(auth.AuthorizationError):
      service.get_async(1).get_result()

  ################################### CANCEL ###################################

  def test_cancel(self):
    bundle = test_util.build_bundle(id=1)
    bundle.put()
    build = service.cancel_async(1, summary_markdown='nope').get_result()
    self.assertEqual(build.proto.status, common_pb2.CANCELED)
    self.assertEqual(build.proto.end_time.ToDatetime(), utils.utcnow())
    self.assertEqual(build.proto.summary_markdown, 'nope')
    self.assertEqual(build.proto.canceled_by, self.current_identity.to_bytes())
    self.assertEqual(build.status_changed_time, utils.utcnow())

    args, _ = swarming.cancel_task_transactionally_async.call_args
    self.assertIsInstance(args[0], model.Build)
    self.assertEqual(args[1], test_util.BUILD_DEFAULTS.infra.swarming)

  def test_cancel_is_idempotent(self):
    build = self.classic_build(id=1)
    build.put()
    service.cancel_async(1).get_result()
    service.cancel_async(1).get_result()

  def test_cancel_started_build(self):
    self.new_started_build(id=1).put()
    service.cancel_async(1).get_result()

  def test_cancel_nonexistent_build(self):
    with self.assertRaises(errors.BuildNotFoundError):
      service.cancel_async(1).get_result()

  def test_cancel_with_auth_error(self):
    self.new_started_build(id=1)
    self.mock_no_perm(user.PERM_BUILDS_CANCEL)
    with self.assertRaises(auth.AuthorizationError):
      service.cancel_async(1).get_result()

  def test_cancel_completed_build(self):

    def test_build_with_status(build_id, status):
      self.classic_build(id=build_id, status=status).put()
      result_build = service.cancel_async(build_id).get_result()
      self.assertEqual(result_build.proto.id, build_id)
      # The status should not change when cancelling completed build
      self.assertEqual(result_build.proto.status, status)

    test_build_with_status(build_id=1, status=common_pb2.SUCCESS)
    test_build_with_status(build_id=2, status=common_pb2.FAILURE)
    test_build_with_status(build_id=3, status=common_pb2.CANCELED)

  def test_cancel_result_details(self):
    self.classic_build(id=1).put()
    result_details = {'message': 'bye bye build'}
    build = service.cancel_async(1, result_details=result_details).get_result()
    self.assertEqual(build.result_details, result_details)

  def test_peek(self):
    build = self.classic_build()
    build.put()
    builds, _ = service.peek(bucket_ids=[build.bucket_id])
    self.assertEqual(builds, [build])

  def test_peek_multi(self):
    build1 = self.classic_build(
        id=1,
        builder=dict(project='chromium', bucket='try'),
    )
    build2 = self.classic_build(
        id=2,
        builder=dict(project='chromium', bucket='try'),
    )
    assert build1.bucket_id == build2.bucket_id
    ndb.put_multi([build1, build2])
    builds, _ = service.peek(bucket_ids=['chromium/try'])
    self.assertEqual(builds, [build2, build1])

  def test_peek_with_paging(self):
    self.put_many_builds(builder=dict(project='chromium', bucket='try'))
    first_page, next_cursor = service.peek(
        bucket_ids=['chromium/try'], max_builds=10
    )
    self.assertTrue(first_page)
    self.assertTrue(next_cursor)

    second_page, _ = service.peek(
        bucket_ids=['chromium/try'], start_cursor=next_cursor
    )

    self.assertTrue(all(b not in second_page for b in first_page))

  def test_peek_with_bad_cursor(self):
    self.put_many_builds(builder=dict(project='chromium', bucket='try'))
    with self.assertRaises(errors.InvalidInputError):
      service.peek(bucket_ids=['chromium/try'], start_cursor='abc')

  def test_peek_without_buckets(self):
    with self.assertRaises(errors.InvalidInputError):
      service.peek(bucket_ids=[])

  def test_peek_with_auth_error(self):
    self.mock_no_perm(user.PERM_BUILDS_LIST)
    build = self.classic_build(builder=dict(project='chromium', bucket='try'))
    build.put()
    with self.assertRaises(auth.AuthorizationError):
      service.peek(bucket_ids=['chromium/try'])

  def test_peek_does_not_return_leased_builds(self):
    self.new_leased_build(builder=dict(project='chromium', bucket='try'))
    builds, _ = service.peek(['chromium/try'])
    self.assertFalse(builds)

  def test_peek_stale_and_lease(self):
    # This is a regression test for crbug.com/1191014
    #
    # The Go version of buildbucket didn't update the legacy status field, which
    # resulted in peek returning builds which weren't actually eligible for
    # leasing.
    build = self.classic_build(status=common_pb2.CANCELED)
    build.put()
    build_id = build.key.id()

    # start with a fresh Build
    build = model.Build.get_by_id(build_id)

    # apply ndb black magicks
    build._clone_properties()  # decouples Build instance from Build._properties
    build._properties['status'] = msgprop.EnumProperty(
        model.BuildStatus, 'status'
    )
    build._properties['status']._code_name = 'status'
    build._properties['status']._set_value(build, model.BuildStatus.SCHEDULED)

    # disable pre_put_hook for this put()
    normalPrePut = model.Build._pre_put_hook
    try:
      model.Build._pre_put_hook = lambda self: None
      build.put()
    finally:
      model.Build._pre_put_hook = normalPrePut

    # Now the build should show up in peek, even though it's canceled.
    builds, _ = service.peek(bucket_ids=[build.bucket_id])
    self.assertEqual(builds, [build])

    # Lease will fix the build status as a side effect
    success, build = service.lease(build_id)
    self.assertFalse(success)

    # So it no longer shows up in peek.
    builds, _ = service.peek(bucket_ids=[build.bucket_id])
    self.assertEqual(builds, [])

  #################################### LEASE ###################################

  def lease(self, build_id, lease_expiration_date=None, expect_success=True):
    success, build = service.lease(
        build_id,
        lease_expiration_date=lease_expiration_date,
    )
    self.assertEqual(success, expect_success)
    return build

  def new_leased_build(self, **build_proto_fields):
    build = self.classic_build(**build_proto_fields)
    build.put()
    return self.lease(build.key.id())

  def test_lease(self):
    expiration_date = utils.utcnow() + datetime.timedelta(minutes=1)
    self.classic_build(id=1).put()
    build = self.lease(1, lease_expiration_date=expiration_date)
    self.assertTrue(build.is_leased)
    self.assertGreater(build.lease_expiration_date, utils.utcnow())
    self.assertEqual(build.leasee, self.current_identity)

  def test_lease_build_with_auth_error(self):
    self.mock_no_perm(user.PERM_BUILDS_LEASE)
    self.classic_build(id=1).put()
    with self.assertRaises(auth.AuthorizationError):
      self.lease(1)

  def test_cannot_lease_a_leased_build(self):
    self.new_leased_build(id=1)
    build = ndb.Key('Build', 1).get()
    self.lease(1, expect_success=False)
    after_build = ndb.Key('Build', 1).get()
    # make sure the NACK lease didn't change the build.
    self.assertEqual(build, after_build)

  def test_cannot_lease_a_nonexistent_build(self):
    with self.assertRaises(errors.BuildNotFoundError):
      service.lease(build_id=42)

  def test_cannot_lease_completed_build(self):
    build = self.classic_build(id=1, status=common_pb2.SUCCESS)
    build.put()
    self.lease(1, expect_success=False)

  def test_cannot_lease_luci_build(self):
    build = test_util.build(id=1)
    build.put()
    with self.assertRaises(errors.InvalidInputError):
      self.lease(1)

  #################################### START ###################################

  def test_validate_malformed_url(self):
    with self.assertRaises(errors.InvalidInputError):
      service.validate_url('svn://sdfsf')

  def test_validate_relative_url(self):
    with self.assertRaises(errors.InvalidInputError):
      service.validate_url('sdfsf')

  def test_validate_nonstring_url(self):
    with self.assertRaises(errors.InvalidInputError):
      service.validate_url(123)

  def start(self, build, url=None, lease_key=None):
    return service.start(build.key.id(), lease_key or build.lease_key, url)

  def test_start(self):
    build = self.new_leased_build()
    build = self.start(build, url='http://localhost')
    self.assertEqual(build.proto.status, common_pb2.STARTED)
    self.assertEqual(build.url, 'http://localhost')
    self.assertEqual(build.proto.start_time.ToDatetime(), self.now)

  def test_start_started_build(self):
    build = self.new_leased_build(id=1)
    lease_key = build.lease_key
    url = 'http://localhost/'

    service.start(1, lease_key, url)
    service.start(1, lease_key, url)
    service.start(1, lease_key, url + '1')

  def test_start_non_leased_build(self):
    self.classic_build(id=1).put()
    with self.assertRaises(errors.LeaseExpiredError):
      service.start(1, 42, None)

  def test_start_completed_build(self):
    self.classic_build(id=1, status=common_pb2.SUCCESS).put()
    with self.assertRaises(errors.BuildIsCompletedError):
      service.start(1, 42, None)

  def test_start_without_lease_key(self):
    with self.assertRaises(errors.InvalidInputError):
      service.start(1, None, None)

  @contextlib.contextmanager
  def callback_test(self, build):
    with mock.patch('notifications.enqueue_notifications_async', autospec=True):
      notifications.enqueue_notifications_async.return_value = future(None)
      build.pubsub_callback = model.PubSubCallback(
          topic='projects/example/topics/buildbucket',
          user_data='hello',
          auth_token='secret',
      )
      build.put()
      yield
      build = build.key.get()
      notifications.enqueue_notifications_async.assert_called_with(build)

  def test_start_creates_notification_task(self):
    build = self.new_leased_build()
    with self.callback_test(build):
      self.start(build)

  ################################## HEARTBEAT #################################

  def test_heartbeat(self):
    build = self.new_leased_build(id=1)
    new_expiration_date = utils.utcnow() + datetime.timedelta(minutes=1)
    build = service.heartbeat(
        1, build.lease_key, lease_expiration_date=new_expiration_date
    )
    self.assertEqual(build.lease_expiration_date, new_expiration_date)

  def test_heartbeat_completed(self):
    self.classic_build(id=1, status=common_pb2.CANCELED).put()
    new_expiration_date = utils.utcnow() + datetime.timedelta(minutes=1)
    with self.assertRaises(errors.BuildIsCompletedError):
      service.heartbeat(1, 0, lease_expiration_date=new_expiration_date)

  def test_heartbeat_timeout(self):
    build = self.classic_build(
        id=1,
        status=common_pb2.INFRA_FAILURE,
        status_details=dict(timeout=dict()),
    )
    build.put()

    new_expiration_date = utils.utcnow() + datetime.timedelta(minutes=1)
    exc_regex = (
        'Build was marked as timed out '
        'because it did not complete for 2 days'
    )
    with self.assertRaisesRegexp(errors.BuildIsCompletedError, exc_regex):
      service.heartbeat(1, 0, lease_expiration_date=new_expiration_date)

  def test_heartbeat_batch(self):
    build = self.new_leased_build(id=1)
    new_expiration_date = utils.utcnow() + datetime.timedelta(minutes=1)
    results = service.heartbeat_batch([
        {
            'build_id': 1,
            'lease_key': build.lease_key,
            'lease_expiration_date': new_expiration_date,
        },
        {
            'build_id': 2,
            'lease_key': 42,
            'lease_expiration_date': new_expiration_date,
        },
    ])

    self.assertEqual(len(results), 2)

    build = build.key.get()
    self.assertEqual(results[0], (1, build, None))

    self.assertIsNone(results[1][1])
    self.assertTrue(isinstance(results[1][2], errors.BuildNotFoundError))

  def test_heartbeat_without_expiration_date(self):
    build = self.new_leased_build(id=1)
    with self.assertRaises(errors.InvalidInputError):
      service.heartbeat(1, build.lease_key, lease_expiration_date=None)

  ################################### COMPLETE #################################

  def new_started_build(self, **build_proto_fields):
    build = self.new_leased_build(**build_proto_fields)
    build = self.start(build)
    return build

  def succeed(self, build, **kwargs):
    return service.succeed(build.key.id(), build.lease_key, **kwargs)

  def test_succeed(self):
    build = self.new_started_build()
    build = self.succeed(build, result_details={'properties': {'foo': 'bar'}})
    self.assertEqual(build.proto.status, common_pb2.SUCCESS)
    self.assertEqual(build.status_changed_time, utils.utcnow())
    self.assertTrue(build.proto.HasField('end_time'))

    out_props = model.BuildOutputProperties.key_for(build.key).get()
    self.assertEqual(test_util.msg_to_dict(out_props.parse()), {'foo': 'bar'})

  def test_succeed_failed(self):
    build = self.classic_build(id=1, status=common_pb2.FAILURE)
    build.put()
    with self.assertRaises(errors.BuildIsCompletedError):
      service.succeed(1, 42)

  def test_succeed_is_idempotent(self):
    build = self.new_started_build(id=1)
    service.succeed(1, build.lease_key)
    service.succeed(1, build.lease_key)

  def test_succeed_with_new_tags(self):
    build = self.new_started_build(id=1, tags=[dict(key='a', value='1')])
    build = self.succeed(build, new_tags=['b:2'])
    self.assertIn('a:1', build.tags)
    self.assertIn('b:2', build.tags)

  def test_fail(self):
    build = self.new_started_build(id=1)
    build = service.fail(1, build.lease_key)
    self.assertEqual(build.proto.status, common_pb2.FAILURE)
    self.assertEqual(build.status_changed_time, utils.utcnow())

  def test_infra_fail(self):
    build = self.new_started_build(id=1)
    build = service.fail(
        1, build.lease_key, failure_reason=model.FailureReason.INFRA_FAILURE
    )
    self.assertEqual(build.proto.status, common_pb2.INFRA_FAILURE)

  def test_fail_with_details(self):
    build = self.new_started_build(id=1)
    result_details = {'transient_failure': True}
    build = service.fail(1, build.lease_key, result_details=result_details)
    self.assertEqual(build.result_details, result_details)

  def test_complete_with_url(self):
    build = self.new_started_build(id=1)
    url = 'http://localhost/1'
    build = self.succeed(build, url=url)
    self.assertEqual(build.url, url)

  def test_complete_not_started_build(self):
    build = self.new_leased_build()
    self.succeed(build)

  def test_completion_creates_notification_task(self):
    build = self.new_started_build()
    with self.callback_test(build):
      self.succeed(build)

  ################################ PAUSE BUCKET ################################

  def test_pause_bucket(self):
    config.put_bucket(
        'chromium',
        'a' * 40,
        test_util.parse_bucket_cfg('name: "master.foo"'),
    )
    config.put_bucket(
        'chromium',
        'a' * 40,
        test_util.parse_bucket_cfg('name: "master.bar"'),
    )

    self.put_many_builds(
        5, builder=dict(project='chromium', bucket='master.foo')
    )
    self.put_many_builds(
        5, builder=dict(project='chromium', bucket='master.bar')
    )

    service.pause('chromium/master.foo', True)
    builds, _ = service.peek(['chromium/master.foo', 'chromium/master.bar'])
    self.assertEqual(len(builds), 5)
    self.assertTrue(all(b.bucket_id == 'chromium/master.bar' for b in builds))

  def test_pause_all_requested_buckets(self):
    config.put_bucket(
        'chromium',
        'a' * 40,
        test_util.parse_bucket_cfg('name: "master.foo"'),
    )
    self.put_many_builds(
        5, builder=dict(project='chromium', bucket='master.foo')
    )

    service.pause('chromium/master.foo', True)
    builds, _ = service.peek(['chromium/master.foo'])
    self.assertEqual(len(builds), 0)

  def test_pause_then_unpause(self):
    build = self.classic_build(builder=dict(project='chromium', bucket='try'))
    build.put()

    config.put_bucket(
        'chromium',
        'a' * 40,
        test_util.parse_bucket_cfg('name: "ci"'),
    )

    service.pause(build.bucket_id, True)
    service.pause(build.bucket_id, True)  # Again, to cover equality case.
    builds, _ = service.peek([build.bucket_id])
    self.assertEqual(len(builds), 0)

    service.pause(build.bucket_id, False)
    builds, _ = service.peek([build.bucket_id])
    self.assertEqual(len(builds), 1)

  def test_pause_bucket_auth_error(self):
    self.mock_no_perm(user.PERM_BUCKETS_PAUSE)
    with self.assertRaises(auth.AuthorizationError):
      service.pause('chromium/no.such.bucket', True)

  def test_pause_invalid_bucket(self):
    config.get_bucket_async.return_value = future((None, None))
    with self.assertRaises(errors.InvalidInputError):
      service.pause('a/#', True)

  def test_pause_luci_bucket(self):
    with self.assertRaises(errors.InvalidInputError):
      service.pause('chromium/luci', True)

  ############################ UNREGISTER BUILDERS #############################

  def test_unregister_builders(self):
    model.Builder(
        id='chromium:try:linux_rel',
        last_scheduled=self.now - datetime.timedelta(weeks=8),
    ).put()
    service.unregister_builders()
    builders = model.Builder.query().fetch()
    self.assertFalse(builders)
