# Copyright 2017 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import (DoesNotRun, DropExpectation, MustRun,
                                        StatusException, StatusSuccess,
                                        StepCommandContains, SummaryMarkdown)

DEPS = [
    'codesearch',
    'depot_tools/bot_update',
    'depot_tools/gclient',
    'recipe_engine/properties',
]


def RunSteps(api):
  api.gclient.set_config('infra')
  api.bot_update.ensure_checkout()
  api.codesearch.set_config(
      api.properties.get('codesearch_config', 'chromium'),
      PROJECT=api.properties.get('project', 'chromium'),
      PLATFORM=api.properties.get('platform', 'linux'),
      SYNC_GENERATED_FILES=api.properties.get('sync_generated_files', True),
      GEN_REPO_BRANCH=api.properties.get('gen_repo_branch', 'main'),
      GEN_REPO_OUT_DIR=api.properties.get('gen_repo_out_dir', ''),
      CORPUS=api.properties.get('corpus', 'chromium-linux'),
  )
  api.codesearch.checkout_generated_files_repo_and_sync({'foo': 'bar'},
                                                        'deadbeef',
                                                        '/path/to/kzip')


def GenTests(api):

  def GetCheckoutAndSyncStepChecks(expected_branch):
    return (
        api.post_process(MustRun, 'git setup'),
        api.post_process(StepCommandContains, 'git fetch', [
            'origin',
            expected_branch,
        ]),
        api.post_process(MustRun, 'git checkout'),
        api.post_process(MustRun, 'read revision'),
        api.post_process(MustRun, 'git clean'),
        api.post_process(MustRun, 'git config'),
        api.post_process(MustRun, 'git config (2)'),
        api.post_process(StepCommandContains, 'sync generated files', [
            '--dest-branch',
            expected_branch,
        ]),
    )

  yield api.test(
      'basic',
      api.properties(buildername='test_buildername', buildnumber=123),
      *GetCheckoutAndSyncStepChecks('main'),
      api.post_process(StatusSuccess),
      api.post_process(DropExpectation),
  )

  yield api.test(
      'specified_branch_and_out_dir',
      api.properties(
          buildername='test_buildername',
          buildnumber=123,
          gen_repo_branch='android',
          gen_repo_out_dir='chromium-android'),
      *GetCheckoutAndSyncStepChecks('android'),
      api.post_process(StatusSuccess),
      api.post_process(DropExpectation),
  )

  yield api.test(
      'false_sync_generated_files',
      api.properties(
          buildername='test_buildername',
          buildnumber=123,
          sync_generated_files=False),
      api.post_process(DoesNotRun, 'git setup'),
      api.post_process(DoesNotRun, 'sync generated files'),
      api.post_process(StatusSuccess),
      api.post_process(DropExpectation),
  )

  yield api.test(
      'generated_repo_not_set_failed',
      api.properties(codesearch_config='base'),
      api.expect_exception('AssertionError'),
      api.post_process(
          SummaryMarkdown,
          "Uncaught Exception: AssertionError('Trying to check out generated "
          "files repo, but the repo is not indicated')"),
      api.post_process(StatusException),
      api.post_process(DropExpectation),
  )
