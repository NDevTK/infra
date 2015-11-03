# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import contextlib
import datetime
import distutils.util
import json
import logging
import os
import re
import shutil
import subprocess
import sys
import tempfile


from infra_libs.time_functions import zulu
from infra.services.master_manager_launcher import desired_state_parser


LOGGER = logging.getLogger(__name__)

MM_REPO = 'https://chrome-internal.googlesource.com/infradata/master-manager'


class MasterNotFoundException(Exception):
  pass


def add_argparse_options(parser):
  parser.add_argument(
      'masters', type=str, nargs='+',
      help='Master(s) to restart. "master." prefix can be omitted.')
  parser.add_argument(
      '-m', '--minutes-in-future', default=15, type=int,
      help='how many minutes in the future to schedule the restart. '
           'use 0 for "now." default %(default)d')
  parser.add_argument('-b', '--bug', default=None, type=str,
                      help='Bug containing master restart request.')
  parser.add_argument('-r', '--reviewer', action='append', type=str,
                      help=(
                          'Reviewer to TBR the CL to. If not specified, '
                          'chooses a random reviewer from OWNERS file'))
  parser.add_argument(
      '-f', '--force', action='store_true',
      help='don\'t ask for confirmation, just commit')
  parser.add_argument(
      '-n', '--no-commit', action='store_true',
      help='update the file, but refrain from performing the actual commit')


def get_restart_time(delta):
  """Returns a zulu time string of when to restart a master, now + delta."""
  restart_time = datetime.datetime.utcnow() + delta
  return zulu.to_zulu_string(restart_time)


@contextlib.contextmanager
def get_master_state_checkout():
  target_dir = tempfile.mkdtemp()
  try:
    LOGGER.info('Cloning %s into %s' % (MM_REPO, target_dir))
    subprocess.call(['git', 'clone', MM_REPO, target_dir])
    LOGGER.info('done')
    yield target_dir
  finally:
    shutil.rmtree(target_dir)


def commit(target, masters, reviewers, bug, timestring, delta, force):
  """Commits the local CL via the CQ."""
  desc = 'Restarting master(s) %s\n' % ', '.join(masters)
  if bug:
    desc += '\nBUG=%s' % bug
  if reviewers:
    desc += '\nTBR=%s' % ', '.join(reviewers)
  subprocess.check_call(
      ['git', 'commit', '--all', '--message', desc], cwd=target)

  print
  print 'Restarting the following masters in %d minutes (%s)' % (
      delta.total_seconds() / 60, timestring)
  for master in sorted(masters):
    print '  %s' % master
  print

  print "This will upload a CL for master_manager.git, TBR an owner, and "
  print "commit the CL through the CQ."
  print

  if not force:
    print 'Commit? [Y/n]:',
    input_string = raw_input()
    if input_string != '' and not distutils.util.strtobool(input_string):
      print 'Aborting.'
      return

  print 'To cancel, edit desired_master_state.json in %s.' % MM_REPO
  print

  LOGGER.info('Uploading to Rietveld and CQ.')
  upload_cmd = [
      'git', 'cl', 'upload',
      '-m', desc,
      '-t', desc, # Title becomes the message of CL. TBR and BUG must be there.
      '-c', '-f',
  ]
  if not reviewers:
    upload_cmd.append('--tbr-owners')
  subprocess.check_call(upload_cmd, cwd=target)


def run(masters, delta, reviewers, bug, force, no_commit):
  """Restart all the masters in the list of masters.

    Schedules the restart for now + delta.
  """
  # Step 1: Acquire a clean master state checkout.
  # This repo is too small to consider caching.
  with get_master_state_checkout() as master_state_dir:
    master_state_json = os.path.join(
        master_state_dir, 'desired_master_state.json')
    restart_time = get_restart_time(delta)

    # Step 2: make modifications to the master state json.
    LOGGER.info('Reading %s' % master_state_json)
    with open(master_state_json, 'r') as f:
      desired_master_state = json.load(f)
    LOGGER.info('Loaded')

    # Validate the current master state file.
    try:
      desired_state_parser.validate_desired_master_state(desired_master_state)
    except desired_state_parser.InvalidDesiredMasterState:
      LOGGER.exception("Failed to validate current master state JSON.")
      return 1

    master_states = desired_master_state.get('master_states', {})
    entries = 0
    for master in masters:
      if not master.startswith('master.'):
        master = 'master.%s' % master
      if master not in master_states:
        msg = '%s not found in master state' % master
        LOGGER.error(msg)
        raise MasterNotFoundException(msg)

      master_states.setdefault(master, []).append({
          'desired_state': 'running', 'transition_time_utc': restart_time
      })
      entries += 1

    LOGGER.info('Writing back to JSON file, %d new entries' % (entries,))
    with open(master_state_json, 'w') as f:
      json.dump(
          desired_master_state, f, sort_keys=True, indent=2,
          separators=(',', ':'))

    # Step 3: Send the patch to Rietveld and commit it via the CQ.
    if no_commit:
      LOGGER.info('Refraining from committing back to repository (--no-commit)')
      return 0

    LOGGER.info('Committing back into repository')
    commit(
        master_state_dir, masters, reviewers, bug, restart_time, delta, force)
