# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import collections
import datetime
import json
import logging
import operator
import os
import re
import urllib
import urlparse

import requests

from infra.tools.builder_alerts import string_helpers


CBE_BASE = 'https://chrome-build-extract.appspot.com'

# Unclear if this should be specific to builds.
class BuildCache(object):
  def __init__(self, root_path):
    self.root_path = root_path

  # Could be in operator.
  def has(self, key):
    path = os.path.join(self.root_path, key)
    return os.path.exists(path)

  def key_age(self, key):
    path = os.path.join(self.root_path, key)
    return datetime.datetime.fromtimestamp(os.path.getmtime(path))

  # Could be attr getter.
  def get(self, key):
    if not self.has(key):
      return None
    path = os.path.join(self.root_path, key)
    with open(path) as cached:
      return json.load(cached)

  # Could be attr setter.
  def set(self, key, json_object):
    path = os.path.join(self.root_path, key)
    cache_dir = os.path.dirname(path)
    if not os.path.exists(cache_dir):
      os.makedirs(cache_dir)
    with open(path, 'w') as cached:
      cached.write(json.dumps(json_object))


def master_name_from_url(master_url):
  return urlparse.urlparse(master_url).path.split('/')[-1]


def build_url(master_url, builder_name, build_number):
  quoted_name = urllib.pathname2url(builder_name)
  args = (master_url, quoted_name, build_number)
  return "%s/builders/%s/builds/%s" % args


def cache_key_for_build(master_url, builder_name, build_number):
  master_name = master_name_from_url(master_url)
  return os.path.join(master_name, builder_name, "%s.json" % build_number)


def fetch_master_json(master_url):  # pragma: no cover
  master_name = master_name_from_url(master_url)
  url = '%s/get_master/%s' % (CBE_BASE, master_name)
  try:
    return requests.get(url).json()
  except ValueError:
    logging.critical('Failed to parse master json file from %s.' % url)
    return {}


def prefill_builds_cache(cache, master_url, builder_name):  # pragma: no cover
  master_name = master_name_from_url(master_url)
  builds_url = '%s/get_builds' % CBE_BASE
  params = { 'master': master_name, 'builder': builder_name }
  response = requests.get(builds_url, params=params)
  builds = []
  try:
    builds = response.json()['builds']
  except ValueError:
    logging.critical(
        'Failed to parse JSON response from master: %s, builder: %s' % (
            master_name, builder_name))
  for build in builds:
    if not build.get('number'):
      index = builds.index(build)
      logging.error('build at index %s in %s missing number?' % (index,
          response.url))
      continue
    build_number = build['number']
    key = cache_key_for_build(master_url, builder_name, build_number)
    cache.set(key, build)
  build_numbers = map(operator.itemgetter('number'), builds)
  logging.debug('Prefilled (%.1fs) %s for %s %s' %
      (response.elapsed.total_seconds(),
      string_helpers.re_range(build_numbers),
      master_name, builder_name))
  return build_numbers


def fetch_and_cache_build(cache, url, cache_key):  # pragma: no cover
  response = requests.get(url)
  if response.status_code != 200:
    logging.error('Failed (%.1fs, %s) %s' % (response.elapsed.total_seconds(),
        response.status_code, response.url))
    return None

  try:
    build = response.json()
  except ValueError, e:
    logging.error('Not caching invalid json: %s (%s): %s\n%s' % (url,
        response.status_code, e, response.text))
    return None

  logging.debug('Fetched (%.1fs) %s' % (response.elapsed.total_seconds(), url))
  cache.set(cache_key, build)
  return build


def is_in_progress(build):
  return build.get('results', None) is None


# "line too long" pylint: disable=C0301
def fetch_build_json(cache, master_url, builder_name, build_number):  # pragma: no cover
  cache_key = cache_key_for_build(master_url, builder_name, build_number)
  build = cache.get(cache_key)
  master_name = master_name_from_url(master_url)

  # We will cache in-progress builds, but only for 2 minutes.
  if build and is_in_progress(build):
    cache_age = datetime.datetime.now() - cache.key_age(cache_key)
    # Round for display.
    cache_age = datetime.timedelta(seconds=round(cache_age.total_seconds()))
    if cache_age.total_seconds() < 120:
      return build
    logging.debug('Expired (%s) %s %s %s' % (cache_age,
        master_name, builder_name, build_number))
    build = None

  if not build:
    cbe_url = ('https://chrome-build-extract.appspot.com/p/%s/builders/'
        '%s/builds/%s?json=1') % (master_name, builder_name, build_number)
    build = fetch_and_cache_build(cache, cbe_url, cache_key)

  if not build:
    buildbot_url = ('https://build.chromium.org/p/%s/json/builders/'
        '%s/builds/%s') % (master_name, builder_name, build_number)
    build = fetch_and_cache_build(cache, buildbot_url, cache_key)

  return build


# This effectively extracts the 'configuration' of the build
# we could extend this beyond repo versions in the future.
def revisions_from_build(build_json):

  def _revision(build_json, property_name):
    for prop_tuple in build_json['properties']:
      if prop_tuple[0] == property_name:
        value = prop_tuple[1]
        # refs/heads/master@{#291569}
        match = re.search('@\{\#(\d+)\}', value)
        return int(match.group(1))

  # TODO(ojan): Use the git hashes. That involves fixing all the code
  # that does arithmetic on revisions in analysis.py.
  REVISION_VARIABLES = [
    ('chromium', 'got_revision_cp'),
    ('blink', 'got_webkit_revision_cp'),
    ('v8', 'got_v8_revision_cp'),
    ('nacl', 'got_nacl_revision_cp'),
    # Skia, for whatever reason, isn't exposed in the buildbot properties so
    # don't bother to include it here.
  ]

  revisions = {}
  for repo_name, buildbot_property in REVISION_VARIABLES:
    revisions[repo_name] = _revision(build_json, buildbot_property)
  return revisions

def latest_update_time_for_builder(last_build):
  last_update = None
  if last_build['times'][1] != None:
    last_update = float(last_build['times'][1])
  else:
    for step in last_build['steps']:
      # A None value for the first step time means the step hasn't started yet.
      # A None value for the second step time means it hasn't finished yet.
      step_time = step['times'][1] or step['times'][0]
      if step_time:
        last_update = max(float(step_time), last_update)
  return last_update


# "line too long" pylint: disable=C0301
def latest_builder_info_for_master(cache, master_url, master_json):
  latest_builder_info = collections.defaultdict(dict)
  master_name = master_name_from_url(master_url)
  for builder_name, builder_json in master_json['builders'].items():
    # recent_builds can include current builds
    recent_builds = set(builder_json['cachedBuilds'])
    active_builds = set(builder_json['currentBuilds'])
    last_finished_id = sorted(recent_builds - active_builds, reverse=True)[0]
    last_build = fetch_build_json(cache,
        master_url, builder_name, last_finished_id)
    latest_builder_info[master_name][builder_name] = {
      'revisions': revisions_from_build(last_build),
      'state': builder_json['state'],
      'lastUpdateTime': latest_update_time_for_builder(last_build),
    }
  return latest_builder_info


def warm_build_cache(cache, master_url, builder_name,
    recent_build_ids, active_builds):  # pragma: no cover
  # Cache active (in-progress) builds:
  match_builder_name = lambda build: build['builderName'] == builder_name
  actives = filter(match_builder_name, active_builds)
  for build in actives:
    key = cache_key_for_build(master_url, builder_name, build['number'])
    cache.set(key, build)

  active_build_ids = [b['number'] for b in active_builds]
  # recent_build_ids includes active ones.
  finished_build_ids = [b for b in recent_build_ids
      if b not in active_build_ids]
  last_build_id = max(finished_build_ids)
  cache_key = cache_key_for_build(master_url, builder_name, last_build_id)

  # We cache in-progress builds, so if the first finished build has a non-None
  # eta, then it's just the cached version from when it was in progress.
  cached_build = cache.get(cache_key)
  if not cached_build or cached_build.get('eta') is not None:
    # reason = 'in progress' if cached_build else 'missing'
    # logging.debug('prefill reason: %s %s' % (max(finished_build_ids), reason))
    prefill_builds_cache(cache, master_url, builder_name)
