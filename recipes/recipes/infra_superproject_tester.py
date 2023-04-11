# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PYTHON_VERSION_COMPATIBILITY = "PY2+3"

DEPS = [
    'recipe_engine/buildbucket',
    'recipe_engine/swarming',
]


def RunSteps(api):
  """Runs infra{_internal} builds for infra_superproject changes."""

  reqs = []
  reqs.append(
      api.buildbucket.schedule_request(
          'infra-internal-try-frontend',
          project='infra_internal',
          bucket='try',
          swarming_parent_run_id=api.swarming.task_id,
      ))

  reqs.append(
      api.buildbucket.schedule_request(
          'infra-internal-tester-bionic-64',
          project='infra_internal',
          bucket='try',
          swarming_parent_run_id=api.swarming.task_id,
      ))

  # Note: infra_superproject is in the infra_internal LUCI project and
  # we cannot assign a task from a different swarming instance as the
  # swarming_parent_run_id.
  reqs.append(
      api.buildbucket.schedule_request(
          'infra-try-bionic-64',
          project='infra',
          bucket='try',
      ))

  reqs.append(
      api.buildbucket.schedule_request(
          'infra-try-mac',
          project='infra',
          bucket='try',
      ))

  reqs.append(
      api.buildbucket.schedule_request(
          'infra-try-win',
          project='infra',
          bucket='try',
      ))

  reqs.append(
      api.buildbucket.schedule_request(
          'infra-try-frontend',
          project='infra',
          bucket='try',
      ))

  api.buildbucket.run(
      reqs, raise_if_unsuccessful=True, step_name='schedule builds')


def GenTests(api):
  yield (api.test('basic', status='INFRA_FAILURE') +
         api.buildbucket.try_build(project='infra/infra_superproject'))
