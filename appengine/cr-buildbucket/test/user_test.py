# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from components import auth
from testing_utils import testing
import mock

from proto.config import project_config_pb2
from test.test_util import future
import config
import errors
import model
import user
import v2

# Shortcuts
Bucket = project_config_pb2.Bucket
Acl = project_config_pb2.Acl


class AclTest(testing.AppengineTestCase):

  def setUp(self):
    super(AclTest, self).setUp()
    self.current_identity = auth.Identity.from_bytes('user:a@example.com')
    self.patch(
        'components.auth.get_current_identity',
        autospec=True,
        return_value=self.current_identity
    )

    self.patch('components.auth.is_admin', autospec=True, return_value=False)

    bucket_a = Bucket(
        name='a',
        acls=[
            Acl(role=Acl.WRITER, group='a-writers'),
            Acl(role=Acl.READER, group='a-readers'),
        ]
    )
    bucket_b = Bucket(
        name='b',
        acls=[
            Acl(role=Acl.WRITER, group='b-writers'),
            Acl(role=Acl.READER, group='b-readers'),
        ]
    )
    bucket_c = Bucket(
        name='c',
        acls=[
            Acl(role=Acl.READER, group='c-readers'),
            Acl(role=Acl.READER, identity='user:a@example.com'),
            Acl(role=Acl.WRITER, group='c-writers'),
        ]
    )
    all_buckets = [bucket_a, bucket_b, bucket_c]
    self.patch(
        'config.get_buckets_async',
        autospec=True,
        return_value=future(all_buckets)
    )

    bucket_map = {b.name: b for b in all_buckets}
    self.patch(
        'config.get_bucket_async',
        autospec=True,
        side_effect=lambda name: future(('chromium', bucket_map.get(name)))
    )

  @mock.patch('components.auth.is_admin', autospec=True)
  @mock.patch('components.auth.is_group_member', autospec=True)
  def test_get_role(self, is_group_member, is_admin):
    is_group_member.side_effect = lambda g, _=None: g == 'a-writers'
    is_admin.return_value = False

    get_role = (lambda *args: user.get_role_async(*args).get_result())

    self.assertEqual(get_role('a'), Acl.WRITER)
    self.assertEqual(get_role('b'), None)
    self.assertEqual(get_role('c'), Acl.READER)
    self.assertEqual(get_role('non.existing'), None)

    is_group_member.side_effect = None
    is_group_member.return_value = False
    self.assertEqual(get_role('a'), None)

    is_admin.return_value = True
    self.assertEqual(get_role('a'), Acl.WRITER)
    self.assertEqual(get_role('non.existing'), None)

  @mock.patch('components.auth.is_admin', autospec=True)
  @mock.patch('components.auth.is_group_member', autospec=True)
  def test_has_any_of_roles(self, is_group_member, is_admin):
    is_group_member.side_effect = lambda g, _=None: g == 'a-readers'
    is_admin.return_value = False

    has_any_of_roles = (
        lambda *args: user.has_any_of_roles_async(*args).get_result()
    )

    self.assertTrue(has_any_of_roles('a', [Acl.READER]))
    self.assertTrue(has_any_of_roles('a', [Acl.READER, Acl.WRITER]))
    self.assertFalse(has_any_of_roles('a', [Acl.WRITER]))
    self.assertFalse(has_any_of_roles('a', [Acl.WRITER, Acl.SCHEDULER]))
    self.assertFalse(has_any_of_roles('b', [Acl.READER]))
    self.assertTrue(has_any_of_roles('c', [Acl.READER]))
    self.assertFalse(has_any_of_roles('c', [Acl.WRITER]))
    self.assertFalse(has_any_of_roles('non.existing', [Acl.READER]))

    is_group_member.side_effect = None
    is_group_member.return_value = False
    self.assertFalse(has_any_of_roles('a', Acl.Role.values()))

    is_admin.return_value = True
    self.assertTrue(has_any_of_roles('a', [Acl.WRITER]))
    self.assertFalse(has_any_of_roles('non-existing', [Acl.WRITER]))

  @mock.patch('components.auth.is_admin', autospec=True)
  @mock.patch('components.auth.is_group_member', autospec=True)
  def test_get_acessible_buckets_async(self, is_group_member, is_admin):
    is_group_member.side_effect = lambda g, _=None: g in ('xxx', 'yyy')
    is_admin.return_value = False

    config.get_buckets_async.return_value = future([
        Bucket(
            name='available_bucket1',
            acls=[
                Acl(role=Acl.READER, group='xxx'),
                Acl(role=Acl.WRITER, group='yyy')
            ],
        ),
        Bucket(
            name='available_bucket2',
            acls=[
                Acl(role=Acl.READER, group='xxx'),
                Acl(role=Acl.WRITER, group='zzz')
            ],
        ),
        Bucket(
            name='available_bucket3',
            acls=[
                Acl(role=Acl.READER, identity='user:a@example.com'),
            ],
        ),
        Bucket(
            name='not_available_bucket',
            acls=[Acl(role=Acl.WRITER, group='zzz')],
        ),
    ])

    availble_buckets = user.get_acessible_buckets_async().get_result()
    # memcache coverage.
    availble_buckets = user.get_acessible_buckets_async().get_result()
    self.assertEqual(
        availble_buckets,
        {'available_bucket1', 'available_bucket2', 'available_bucket3'}
    )

    is_admin.return_value = True
    self.assertIsNone(user.get_acessible_buckets_async().get_result())

  def mock_has_any_of_roles(self, current_identity_roles):
    current_identity_roles = set(current_identity_roles)

    def has_any_of_roles_async(_bucket, roles):
      return future(current_identity_roles.intersection(roles))

    self.patch(
        'user.has_any_of_roles_async', side_effect=has_any_of_roles_async
    )

  def test_can(self):
    self.mock_has_any_of_roles([Acl.READER])
    self.assertTrue(user.can('bucket', user.Action.VIEW_BUILD))
    self.assertFalse(user.can('bucket', user.Action.CANCEL_BUILD))
    self.assertFalse(user.can('bucket', user.Action.WRITE_ACL))

    # Memcache coverage
    self.assertFalse(user.can('bucket', user.Action.WRITE_ACL))
    self.assertFalse(user.can_add_build_async('bucket').get_result())

  def test_can_no_roles(self):
    self.mock_has_any_of_roles([])
    for action in user.Action:
      self.assertFalse(user.can('bucket', action))

  def test_can_bad_input(self):
    with self.assertRaises(errors.InvalidInputError):
      user.can('bad bucket name', user.Action.VIEW_BUILD)

  def test_can_view_build(self):
    self.mock_has_any_of_roles([Acl.READER])
    build = model.Build(bucket='bucket')
    self.assertTrue(user.can_view_build(build))
    self.assertFalse(user.can_lease_build(build))
