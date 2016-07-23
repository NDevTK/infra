# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Rolls recipes.cfg dependencies for public projects."""

DEPS = [
  'recipe_autoroller',

  'build/luci_config',

  'recipe_engine/properties',
  'recipe_engine/raw_io',
  'recipe_engine/step',
]

from recipe_engine import recipe_api


# Toposorted for best results.
# TODO(phajdan.jr): get the list of public projects from luci-config.
PROJECTS = [
  'depot_tools',
  'build',
  'infra',
]


PROPERTIES = {
  'projects': recipe_api.Property(default=PROJECTS),

  # To generate an auth token for running locally, run
  #   infra/go/bin/authutil login
  'auth_token': recipe_api.Property(default=None),
}


def get_auth_token(api):
  """Get an auth token from the recipe-roller service account."""

  result = api.step('Get auth token',
      ['/opt/infra-tools/authutil', 'token',
       '-service-account-json='
       '/creds/service_accounts/service-account-recipe-roller.json'],
      stdout=api.raw_io.output(),
      step_test_data=lambda: api.raw_io.test_api.stream_output('ya29.foobar'))
  return result.stdout.strip()


def RunSteps(api, projects, auth_token):
  api.recipe_autoroller.prepare_checkout()

  if not auth_token:
    auth_token = get_auth_token(api)
  api.luci_config.c.auth_token = auth_token

  api.recipe_autoroller.roll_projects(projects)



def GenTests(api):
  yield (
      api.test('basic') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.recipe_autoroller.roll_data('build') +
      api.recipe_autoroller.new_upload('build')
  )

  yield (
      api.test('nontrivial') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.recipe_autoroller.roll_data('build', trivial=False) +
      api.recipe_autoroller.new_upload('build')
  )

  yield (
      api.test('empty') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.recipe_autoroller.roll_data('build', empty=True)
  )

  yield (
      api.test('failure') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.recipe_autoroller.roll_data('build', success=False)
  )

  yield (
      api.test('previously_uploaded') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.recipe_autoroller.roll_data('build') +
      api.recipe_autoroller.previously_uploaded('build')
  )

  yield (
      api.test('failed_upload') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.recipe_autoroller.roll_data('build') +
      api.recipe_autoroller.new_upload('build') +
      api.override_step_data(
          'build.git cl issue',
          api.raw_io.stream_output('Issue number: None (None)'))
  )
