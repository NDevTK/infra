#!/usr/bin/env python3
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import datetime
import json
import os
import urllib.request

_PROJECT_NAME = 'protocolbuffers/txtpbfmt'
_COMMITS_URL = 'https://api.github.com/repos/{}/commits?per_page=1&until={}'
_ARCHIVE_URL = 'https://github.com/{}/archive/{}.tar.gz'


def get_last_commit_of_the_week():
  """Returns data of last commit until the end of last week."""
  # Get a datetime object representing today ignoring time of day.
  # Use UTC to avoid potential date difference due to timezones.
  utc_now = datetime.datetime.now(datetime.timezone.utc)
  utc_today = datetime.datetime.combine(utc_now.date(), datetime.time.min)
  # Find number of days since last monday.
  days_since_week_start = utc_today.weekday()
  # Get time at the start of this week (start of Monday of this week).
  end_of_last_week = utc_today - datetime.timedelta(days=days_since_week_start)
  url = _COMMITS_URL.format(_PROJECT_NAME, end_of_last_week.isoformat())
  # Search for the last commit until the start of this week.
  return json.load(urllib.request.urlopen(url))[0]


def do_latest():
  print(get_last_commit_of_the_week()['sha'])


def get_download_url(sha):
  partial_manifest = {
      'url': [_ARCHIVE_URL.format(_PROJECT_NAME, sha)],
      'ext': '.tar.gz',
  }
  print(json.dumps(partial_manifest))


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers()

  latest = sub.add_parser("latest")
  latest.set_defaults(func=do_latest)

  download = sub.add_parser("get_url")
  download.set_defaults(
      func=lambda: get_download_url(os.environ['_3PP_VERSION']))

  opts = ap.parse_args()
  opts.func()


if __name__ == '__main__':
  main()
