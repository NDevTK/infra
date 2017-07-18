# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.recipe_api import Property

DEPS = [
    'depot_tools/bot_update',
    'depot_tools/gclient',
    'recipe_engine/path',
    'recipe_engine/properties',
    'recipe_engine/python',
    'recipe_engine/step',
]

PROPERTIES = {
  'workflow': Property(
      kind=str, help=('Path to the dataflow workflow you would like to '
                      'execute. Will be appended to the infra checkout path.')),
  'job_name': Property(
      kind=str, help=('Name that appears on the Dataflow console. Must match '
                      'the regular expression [a-z]([-a-z0-9]{0,38}[a-z0-9])')),
}

def RunSteps(api, workflow, job_name):
  api.gclient.set_config('infra')
  api.bot_update.ensure_checkout()
  api.gclient.runhooks()
  workflow_path = api.path['checkout']
  workflow_path = workflow_path.join(*workflow.split('/'))
  setup_path = api.path['checkout'].join('infra', 'dataflow', 'events',
                                         'setup.py')
  args = ['--job_name', job_name, '--project', 'chrome-infra-events',
          '--runner', 'DataflowRunner', '--setup_file', setup_path,
          '--staging_location', 'gs://dataflow-chrome-infra/events/staging',
          '--temp_location', 'gs://dataflow-chrome-infra/events/temp',
          '--save_main_session']
  api.python('Remote execute', workflow_path, args)

def GenTests(api):
  yield api.test('basic') + api.properties(
      workflow='infra/dataflow/events/cq_attempts.py', job_name='cq-attempts')
