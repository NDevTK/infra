# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from components import auth
from parameterized import parameterized
from testing_utils import testing
import mock

from google.appengine.ext import ndb

from go.chromium.org.luci.buildbucket.proto import project_config_pb2
from test.test_util import future
import config
import errors
import user


# Representative subsets of permissions per role.
READER_PERMS = frozenset([user.PERM_BUILDS_GET])
SCHEDULER_PERMS = READER_PERMS | frozenset([user.PERM_BUILDS_ADD])
WRITER_PERMS = SCHEDULER_PERMS | frozenset([user.PERM_BUILDS_LEASE])


class UserTest(testing.AppengineTestCase):

  def setUp(self):
    super(UserTest, self).setUp()
    self.current_identity = auth.Identity.from_bytes('user:a@example.com')
    self.patch(
        'components.auth.get_current_identity',
        autospec=True,
        side_effect=lambda: self.current_identity
    )
    user.clear_request_cache()

    self.patch('components.auth.is_admin', autospec=True, return_value=False)

    self.perms = {}  # realm -> [(group|identity, set of permissions it has)]

    def has_permission(perm, realms):
      assert isinstance(perm, auth.Permission)
      assert isinstance(realms, list)
      caller = auth.get_current_identity().to_bytes()
      for r in realms:
        for principal, granted in self.perms.get(r, []):
          applies = caller == principal or auth.is_group_member(principal)
          if applies and perm in granted:
            return True
      return False

    self.patch(
        'components.auth.has_permission',
        autospec=True,
        side_effect=has_permission
    )

    config.put_bucket('p1', 'ignored-rev', project_config_pb2.Bucket(name='a'))
    self.perms['p1:a'] = [
        ('a-writers', WRITER_PERMS),
        ('a-readers', READER_PERMS),
        ('project:p1', SCHEDULER_PERMS),  # implicit
    ]

    config.put_bucket('p2', 'ignored-rev', project_config_pb2.Bucket(name='b'))
    self.perms['p2:b'] = [
        ('b-writers', WRITER_PERMS),
        ('b-readers', READER_PERMS),
        ('project:p2', SCHEDULER_PERMS),  # implicit
    ]

    config.put_bucket('p3', 'ignored-rev', project_config_pb2.Bucket(name='c'))
    self.perms['p3:c'] = [
        ('c-readers', READER_PERMS),
        ('user:a@example.com', READER_PERMS),
        ('c-writers', WRITER_PERMS),
        ('project:p1', READER_PERMS),
        ('project:p3', SCHEDULER_PERMS),  # implicit
    ]

  @parameterized.expand([
      (user.PERM_BUILDS_GET, ['a-readers'], {'p1/a', 'p3/c'}),
      (user.PERM_BUILDS_ADD, ['b-writers'], {'p2/b'}),
  ])
  @mock.patch('components.auth.is_group_member', autospec=True)
  def test_buckets_by_perm_async(self, perm, groups, expected, is_group_member):
    is_group_member.side_effect = lambda g, _=None: g in groups

    # Cold caches.
    buckets = user.buckets_by_perm_async(perm).get_result()
    self.assertEqual(buckets, expected)

    # Test coverage of ndb.Future caching.
    buckets = user.buckets_by_perm_async(perm).get_result()
    self.assertEqual(buckets, expected)

    # Memcache coverage.
    user.clear_request_cache()
    buckets = user.buckets_by_perm_async(perm).get_result()
    self.assertEqual(buckets, expected)

  @mock.patch('components.auth.is_group_member', autospec=True)
  def test_buckets_by_perm_async_for_project(self, is_group_member):
    is_group_member.side_effect = lambda g, _=None: False
    self.current_identity = auth.Identity.from_bytes('project:p1')

    buckets = user.buckets_by_perm_async(user.PERM_BUILDS_GET).get_result()
    self.assertEqual(
        buckets,
        {
            'p1/a',  # implicit by being in the same project
            'p3/c',  # explicitly set in acls {...}, see setUp()
        }
    )

  @mock.patch('components.auth.is_group_member', autospec=True)
  def test_has_perm(self, is_group_member):
    is_group_member.side_effect = lambda g, _=None: g == 'a-readers'
    self.assertTrue(user.has_perm(user.PERM_BUILDS_GET, 'p1/a'))
    self.assertFalse(user.has_perm(user.PERM_BUILDS_ADD, 'p1/a'))
    self.assertFalse(user.has_perm(user.PERM_BUILDS_GET, 'p2/b'))

    is_group_member.side_effect = lambda g, _=None: g == 'a-writers'
    self.assertTrue(user.has_perm(user.PERM_BUILDS_GET, 'p1/a'))
    self.assertTrue(user.has_perm(user.PERM_BUILDS_ADD, 'p1/a'))
    self.assertFalse(user.has_perm(user.PERM_BUILDS_GET, 'p2/b'))

    is_group_member.return_value = False
    auth.is_admin.return_value = True
    self.assertTrue(user.has_perm(user.PERM_BUILDS_GET, 'p1/a'))
    self.assertTrue(user.has_perm(user.PERM_BUILDS_ADD, 'p1/a'))
    self.assertTrue(user.has_perm(user.PERM_BUILDS_GET, 'p2/b'))

  def test_has_perm_bad_input(self):
    with self.assertRaises(errors.InvalidInputError):
      bid = 'bad project id/bucket'
      user.has_perm(user.PERM_BUILDS_GET, bid)
    with self.assertRaises(errors.InvalidInputError):
      bid = 'project_id/bad bucket name'
      user.has_perm(user.PERM_BUILDS_GET, bid)

  @mock.patch('components.auth.is_group_member', autospec=True)
  def test_filter_buckets_by_perm(self, is_group_member):
    self.current_identity = auth.Identity.from_bytes('user:a@example.com')
    is_group_member.side_effect = lambda g, _=None: g == 'a-readers'

    all_buckets = ['p1/a', 'p2/b', 'p3/c', 'p/unknown']

    filtered = user.filter_buckets_by_perm(user.PERM_BUILDS_GET, all_buckets)
    self.assertEqual(
        filtered,
        {
            'p1/a',  # via a-readers group
            'p3/c',  # via direct identity reference in ACLs
        }
    )

    self.current_identity = auth.Identity.from_bytes('project:p2')
    filtered = user.filter_buckets_by_perm(user.PERM_BUILDS_ADD, all_buckets)
    self.assertEqual(filtered, {'p2/b'})

  def test_parse_identity(self):
    self.assertEqual(
        user.parse_identity('user:a@example.com'),
        auth.Identity('user', 'a@example.com'),
    )
    self.assertEqual(
        auth.Identity('user', 'a@example.com'),
        auth.Identity('user', 'a@example.com'),
    )

    self.assertEqual(
        user.parse_identity('a@example.com'),
        auth.Identity('user', 'a@example.com'),
    )

    with self.assertRaises(errors.InvalidInputError):
      user.parse_identity('a:b')


class GetOrCreateCachedFutureTest(testing.AppengineTestCase):
  maxDiff = None

  def test_unfinished_future_in_different_context(self):
    # This test essentially asserts ndb behavior that we assume in
    # user._get_or_create_cached_future.

    ident1 = auth.Identity.from_bytes('user:1@example.com')
    ident2 = auth.Identity.from_bytes('user:2@example.com')

    # First define a correct async function that uses caching.
    log = []

    @ndb.tasklet
    def compute_async(x):
      log.append('compute_async(%r) started' % x)
      yield ndb.sleep(0.001)
      log.append('compute_async(%r) finishing' % x)
      raise ndb.Return(x)

    def compute_cached_async(x):
      log.append('compute_cached_async(%r)' % x)
      # Use different identities to make sure _get_or_create_cached_future is
      # OK with that.
      ident = ident1 if x % 2 else ident2
      return user._get_or_create_cached_future(
          ident, x, lambda: compute_async(x)
      )

    # Now call compute_cached_async a few times, but stop on the first result,
    # and exit the current ndb context leaving remaining futures unfinished.

    class Error(Exception):
      pass

    with self.assertRaises(Error):
      # This code is intentionally looks realistic.
      futures = [compute_cached_async(x) for x in xrange(5)]
      for f in futures:  # pragma: no branch
        f.get_result()
        log.append('got result')
        # Something bad happened during processing.
        raise Error()

    # Assert that only first compute_async finished.
    self.assertEqual(
        log,
        [
            'compute_cached_async(0)',
            'compute_cached_async(1)',
            'compute_cached_async(2)',
            'compute_cached_async(3)',
            'compute_cached_async(4)',
            'compute_async(0) started',
            'compute_async(1) started',
            'compute_async(2) started',
            'compute_async(3) started',
            'compute_async(4) started',
            'compute_async(0) finishing',
            'got result',
        ],
    )
    log[:] = []

    # Now we assert that waiting for another future, continues execution.
    self.assertEqual(compute_cached_async(3).get_result(), 3)
    self.assertEqual(
        log,
        [
            'compute_cached_async(3)',
            'compute_async(1) finishing',
            'compute_async(2) finishing',
            'compute_async(3) finishing',
        ],
    )
    # Note that compute_async(4) didin't finish.
