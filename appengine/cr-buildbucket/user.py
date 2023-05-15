# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""User-related functions, including access control list implementation."""

import logging
import os
import threading

from google.appengine.ext import ndb

from components import auth
from components import utils

import config
import errors

# Group users to update builds. They are expected to be robots.
UPDATE_BUILD_ALLOWED_USERS = 'buildbucket-update-build-users'


ALL_PERMISSIONS = set()


def _permission(name):
  perm = auth.Permission(name)
  ALL_PERMISSIONS.add(perm)
  return perm


# Builds permissions.

# See all information about a build.
PERM_BUILDS_GET = _permission('buildbucket.builds.get')
# See a limited set of information about a build.
PERM_BUILDS_GET_LIMITED = _permission('buildbucket.builds.getLimited')
# List and search builds in a bucket.
PERM_BUILDS_LIST = _permission('buildbucket.builds.list')
# Schedule new builds in the bucket.
PERM_BUILDS_ADD = _permission('buildbucket.builds.add')
# Cancel a build in the bucket.
PERM_BUILDS_CANCEL = _permission('buildbucket.builds.cancel')
# Lease and control a build via v1 API, deprecated.
PERM_BUILDS_LEASE = _permission('buildbucket.builds.lease')
# Unlease and reset state of an existing build via v1 API, deprecated.
PERM_BUILDS_RESET = _permission('buildbucket.builds.reset')

# Builders permissions.

# See existence and metadata of a builder (but not its builds).
PERM_BUILDERS_GET = _permission('buildbucket.builders.get')
# List and search builders (but not builds).
PERM_BUILDERS_LIST = _permission('buildbucket.builders.list')
# Set the next build number.
PERM_BUILDERS_SET_NUM = _permission('buildbucket.builders.setBuildNumber')

# Bucket permissions.

# See existence of a bucket, used only by v1 APIs, deprecated.
PERM_BUCKETS_GET = _permission('buildbucket.buckets.get')
# Delete all scheduled builds from a bucket.
PERM_BUCKETS_DELETE_BUILDS = _permission('buildbucket.buckets.deleteBuilds')
# Pause/resume leasing builds in a bucket via v1 API, deprecated.
PERM_BUCKETS_PAUSE = _permission('buildbucket.buckets.pause')

# Forbid adding more permission from other modules or tests after this point.
ALL_PERMISSIONS = frozenset(ALL_PERMISSIONS)


@ndb.tasklet
def has_perm_async(perm, bucket_id):
  """Returns True if the caller has the given permission in the bucket.

  Args:
    perm: an instance of auth.Permission.
    bucket_id: a bucket ID string, i.e. "<project>/<bucket>".
  """
  assert isinstance(perm, auth.Permission), perm
  assert perm in ALL_PERMISSIONS, perm
  config.validate_bucket_id(bucket_id)

  # Convert to a realm ID (it uses ':' separator).
  project, bucket = config.parse_bucket_id(bucket_id)
  realm = '%s:%s' % (project, bucket)

  # Check realm ACLs first.
  if auth.has_permission(perm, [realm]):
    raise ndb.Return(True)

  # For compatibility with legacy ALCs, administrators have implicit access to
  # everything. Log when this rule is invoked, since it's surprising and it
  # something we might want to get rid of after everything is migrated to
  # Realms.
  if auth.is_admin():
    logging.warning(
        'ADMIN_FALLBACK: perm=%r bucket=%r caller=%r',
        perm,
        bucket_id,
        auth.get_current_identity().to_bytes(),
    )
    raise ndb.Return(True)

  raise ndb.Return(False)


def has_perm(perm, bucket_id):
  """Returns True if the caller has the given permission in the bucket.

  Args:
    perm: an instance of auth.Permission.
    bucket_id: a bucket ID string, i.e. "<project>/<bucket>".
  """
  return has_perm_async(perm, bucket_id).get_result()


def filter_buckets_by_perm(perm, bucket_ids):
  """Filters given buckets keeping only ones the caller has the permission in.

  Note that this function is not async!

  Args:
    perm: an instance of auth.Permission.
    bucket_ids: an iterable with bucket IDs.

  Returns:
    A set of bucket IDs.
  """
  pairs = utils.async_apply(
      bucket_ids if isinstance(bucket_ids, set) else set(bucket_ids),
      lambda bid: has_perm_async(perm, bid),
      unordered=True,
  )
  return {bid for bid, has_perm in pairs if has_perm}


def buckets_by_perm_async(perm):
  """Returns buckets that the caller has the given permission in.

  Results are memcached for 10 minutes per (identity, perm) pair.

  Args:
    perm: an instance of auth.Permission.

  Returns:
    A set of bucket IDs.
  """
  assert isinstance(perm, auth.Permission), perm
  assert perm in ALL_PERMISSIONS, perm

  identity = auth.get_current_identity()
  identity_str = identity.to_bytes()

  @ndb.tasklet
  def impl():
    ctx = ndb.get_context()
    cache_key = 'buckets_by_perm/%s/%s' % (identity_str, perm)
    matching_buckets = yield ctx.memcache_get(cache_key)
    if matching_buckets is not None:
      raise ndb.Return(matching_buckets)

    logging.info('Computing a set of buckets %r has %r in', identity_str, perm)
    all_buckets = yield config.get_all_bucket_ids_async()
    per_bucket = yield [has_perm_async(perm, bid) for bid in all_buckets]
    matching_buckets = {bid for bid, has in zip(all_buckets, per_bucket) if has}

    # Cache for 10 min
    yield ctx.memcache_set(cache_key, matching_buckets, 10 * 60)
    raise ndb.Return(matching_buckets)

  return _get_or_create_cached_future(
      identity, 'buckets_by_perm/%s' % perm, impl
  )


################################################################################
## Granular actions. API uses these.


@ndb.tasklet
def can_update_build_async():  # pragma: no cover
  """Returns if the current identity is whitelisted to update builds."""
  # TODO(crbug.com/1091604): Implementing using has_perm_async.
  raise ndb.Return(auth.is_group_member(UPDATE_BUILD_ALLOWED_USERS))


################################################################################
## User-related helpers used by other modules.


def current_identity_cannot(action_format, *args):  # pragma: no cover
  """Returns AuthorizationError."""
  action = action_format % args
  msg = 'User %s cannot %s' % (auth.get_current_identity().to_bytes(), action)
  logging.warning(msg)
  return auth.AuthorizationError(msg)


def parse_identity(identity):
  """Parses an identity string if it is a string."""
  if isinstance(identity, basestring):
    if not identity:  # pragma: no cover
      return None
    if ':' not in identity:  # pragma: no branch
      identity = 'user:%s' % identity
    try:
      identity = auth.Identity.from_bytes(identity)
    except ValueError as ex:
      raise errors.InvalidInputError('Invalid identity: %s' % ex)
  return identity


################################################################################
## Implementation.

_thread_local = threading.local()


def _get_or_create_cached_future(identity, key, create_future):
  """Returns a future cached in the current GAE request context.

  Uses the pair (identity, key) as the caching key.

  Using this function may cause RuntimeError with a deadlock if the returned
  future is not waited for before leaving an ndb context, but that's a bug
  in the first place.
  """
  assert isinstance(identity, auth.Identity), identity
  full_key = (identity, key)

  # Docs:
  # https://cloud.google.com/appengine/docs/standard/python/how-requests-are-handled#request-ids
  req_id = os.environ['REQUEST_LOG_ID']
  cache = getattr(_thread_local, 'request_cache', {})
  if cache.get('request_id') != req_id:
    cache = {
        'request_id': req_id,
        'futures': {},
    }
    _thread_local.request_cache = cache

  fut_entry = cache['futures'].get(full_key)
  if fut_entry is None:
    fut_entry = {
        'future': create_future(),
        'ndb_context': ndb.get_context(),
    }
    cache['futures'][full_key] = fut_entry
  assert (
      fut_entry['future'].done() or
      ndb.get_context() is fut_entry['ndb_context']
  )
  return fut_entry['future']


def clear_request_cache():
  _thread_local.request_cache = {}
