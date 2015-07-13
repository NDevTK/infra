# Copyright (c) 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Runs a pipeline to detect suspicious commits in Chromium."""


DEPS = [
  'bot_update',
  'gclient',
  'gsutil',
  'path',
  'properties',
  'python',
]


def RunSteps(api):
  api.gclient.set_config('infra_with_chromium')
  api.bot_update.ensure_checkout(force=True)
  api.gclient.runhooks()
  dirname = api.path.mkdtemp('antibody').join('antibody')

  cmd = ['infra.tools.antibody']
  cmd.extend(['--sql-password-file', '/home/chrome-bot/.antibody_password'])
  cmd.extend(['--git-checkout-path', api.m.path['slave_build'].join('infra')])
  cmd.extend(['--output-dir-path', dirname])
  cmd.extend(['--since', '2015-01-01'])
  cmd.extend(['--run-antibody'])

  api.python('Antibody', 'run.py', cmd,
             cwd=api.m.path['slave_build'].join('infra'))
  api.gsutil(['cp', '-r', '-a', 'public-read', dirname, 'gs://antibody/'])


def GenTests(api):
  yield (api.test('antibody') +
         api.properties(mastername='chromium.infra.cron',
                        buildername='antibody',
                        slavename='fake-slave'))
