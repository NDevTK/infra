# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Rolls recipes.cfg dependencies for public projects."""

DEPS = [
  'recipe_autoroller',
  'recipe_utils',

  'build/luci_config',

  'recipe_engine/properties',
  'recipe_engine/raw_io',
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
}


def RunSteps(api, projects):
  api.recipe_autoroller.roll_projects(projects)


def GenTests(api):
  yield (
      api.test('basic') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.luci_config.get_project_config(
          'build', 'recipes.cfg',
          api.recipe_utils.make_recipe_config('build', [])) +
      api.recipe_autoroller.roll_data('build') +
      api.recipe_autoroller.new_upload('build')
  )

  yield (
      api.test('nontrivial') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.luci_config.get_project_config(
          'build', 'recipes.cfg',
          api.recipe_utils.make_recipe_config('build', [])) +
      api.recipe_autoroller.roll_data('build', trivial=False) +
      api.recipe_autoroller.new_upload('build')
  )

  yield (
      api.test('empty') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.luci_config.get_project_config(
          'build', 'recipes.cfg',
          api.recipe_utils.make_recipe_config('build', [])) +
      api.recipe_autoroller.roll_data('build', empty=True)
  )

  yield (
      api.test('failure') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.luci_config.get_project_config(
          'build', 'recipes.cfg',
          api.recipe_utils.make_recipe_config('build', [])) +
      api.recipe_autoroller.roll_data('build', success=False)
  )

  yield (
      api.test('previously_uploaded') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.luci_config.get_project_config(
          'build', 'recipes.cfg',
          api.recipe_utils.make_recipe_config('build', [])) +
      api.recipe_autoroller.roll_data('build') +
      api.recipe_autoroller.previously_uploaded('build')
  )

  yield (
      api.test('failed_upload') +
      api.properties(projects=['build']) +
      api.luci_config.get_projects(['build']) +
      api.luci_config.get_project_config(
          'build', 'recipes.cfg',
          api.recipe_utils.make_recipe_config('build', [])) +
      api.recipe_autoroller.roll_data('build') +
      api.recipe_autoroller.new_upload('build') +
      api.override_step_data(
          'build.git cl issue',
          api.raw_io.stream_output('Issue number: None (None)'))
  )

  yield (
      api.test('empty_when_dependent_repos_inconsistent') +
      api.properties(projects=['infra']) +
      api.luci_config.get_projects(
          ['recipe_engine', 'depot_tools', 'build', 'infra']) +
      api.luci_config.get_project_config(
          'recipe_engine', 'recipes.cfg',
          api.recipe_utils.make_recipe_config('recipe_engine', [])) +
      api.luci_config.get_project_config(
          'depot_tools', 'recipes.cfg',
          api.recipe_utils.make_recipe_config('depot_tools', [
              'recipe_engine'])) +
      api.luci_config.get_project_config(
          'build', 'recipes.cfg',
          api.recipe_utils.make_recipe_config('build', [
              'recipe_engine', 'depot_tools'])) +
      api.luci_config.get_project_config(
          'infra', 'recipes.cfg',
          api.recipe_utils.make_recipe_config('build', [
              'recipe_engine', 'depot_tools', 'build'])) +
      api.recipe_autoroller.roll_data('infra', empty=True)
  )
