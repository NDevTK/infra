# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.go.chromium.org.luci.buildbucket.proto.common import GerritChange

PYTHON_VERSION_COMPATIBILITY = "PY2+3"

DEPS = [
    'infra_checkout',
    'recipe_engine/buildbucket',
    'recipe_engine/cipd',
    'recipe_engine/context',
    'recipe_engine/file',
    'recipe_engine/nodejs',
    'recipe_engine/path',
    'recipe_engine/platform',
    'recipe_engine/properties',
    'recipe_engine/step',
    'depot_tools/bot_update',
    'depot_tools/gclient',
]

def RunSteps(api):
  assert api.platform.is_linux, 'Unsupported platform, only Linux is supported.'
  cl = api.buildbucket.build.input.gerrit_changes[0]
  project_name = cl.project

  # For builds scheduled for an infra/infra_superproject change,
  # the cl project is infra_superproject, but the builder project
  # should be one of 'infra' or 'infra_internal'.
  if project_name == 'infra/infra_superproject':
    builder_project = api.buildbucket.build.builder.project
    assert builder_project in ('infra', 'infra-internal'), (
        'unknown builder project: "%s" for infra_superproject change' %
        builder_project)
    project_name = project_name[:-len('_superproject')]
    if builder_project == 'infra':
      patch_root = 'infra'
    else:
      patch_root = 'infra_internal'
  else:
    assert project_name in ('infra/infra', 'infra/infra_internal',
                            'infra/luci/luci-go'), ('unknown project: "%s"' %
                                                    project_name)
    patch_root = project_name.split('/')[-1]

  config_name = patch_root.replace("-", "_")
  # TODO(crbug.com/1415507): Remove '_superproject' suffix when
  # migration is complete and configs have been renamed.
  if config_name in ('infra', 'infra_internal'):
    config_name += '_superproject'

  path = api.path['cache'].join('builder')
  api.file.ensure_directory('ensure builder dir', path)

  override_revisions = api.infra_checkout.get_footer_infra_deps_overrides(cl)
  with api.context(cwd=path):
    api.gclient.set_config(config_name)
    api.bot_update.ensure_checkout(
        patch_root=patch_root, recipe_revision_overrides=override_revisions)
    api.gclient.runhooks()

  # Project => (where to find it, how to run its tests).
  checkout_path, runner = {
      'infra/infra': ('infra', RunInfraFrontendTests),
      'infra/infra_internal': ('infra_internal', RunInfraInternalFrontendTests),
      'infra/luci/luci-go': ('go/src/go.chromium.org/luci', RunLuciGoTests),
  }[project_name]
  repo_checkout_root = api.path['checkout'].join(checkout_path)

  # Read the desired nodejs version from <repo>/build/NODEJS_VERSION.
  version = api.file.read_text(
      'read NODEJS_VERSION',
      repo_checkout_root.join('build', 'NODEJS_VERSION'),
      test_data='6.6.6\n',
  ).strip().lower()

  # Bootstrap nodejs at that version and run tests.
  with api.nodejs(version):
    runner(api, repo_checkout_root)


def RunInfraInternalFrontendTests(api, root_path):
  """This function runs UI tests in `infra_internal` project.
  """

  # Add your infra_internal tests here following this example:
  # cwd = api.path['checkout'].join('path', 'to', 'ui', 'root')
  # RunFrontendTests(api, env, cwd, 'myapp')
  # `myapp` is the name that will show up in the step.

  testhaus = root_path.join('go', 'src', 'infra_internal', 'appengine',
                            'testhaus')
  RunFrontendTests(api, testhaus.join('frontend', 'ui'), 'testhaus')

  cwd = root_path.join('go', 'src', 'infra_internal', 'appengine', 'spike',
                       'appengine', 'frontend', 'ui')
  RunFrontendTests(api, cwd, 'spike')


def RunInfraFrontendTests(api, root_path):
  """This function runs the UI tests in `infra` project.
  """

  cwd = root_path.join('appengine', 'monorail')
  RunFrontendTests(api, cwd, 'monorail')

  cwd = root_path.join('go', 'src', 'infra', 'appengine', 'dashboard',
                       'frontend')
  RunFrontendTests(api, cwd, 'chopsdash')


def RunLuciGoTests(api, root_path):
  """This function runs UI tests in the `luci-go` project.
  """

  cwd = root_path.join('analysis', 'frontend', 'ui')
  RunFrontendTests(api, cwd, 'analysis')

  cwd = root_path.join('bisection', 'frontend', 'ui')
  RunFrontendTests(api, cwd, 'bisection')

  cwd = root_path.join('milo', 'ui')
  RunFrontendTests(api, cwd, 'milo')


def RunFrontendTests(api, cwd, app_name):
  with api.context(cwd=cwd):
    api.step(('%s npm install' % app_name), ['npm', 'ci'])
    api.step(('%s test' % app_name), ['npm', 'run', 'test'])


def GenTests(api):
  yield (
      api.test('basic') +
      api.buildbucket.try_build(project='infra/infra'))
  yield (
      api.test('basic-internal') +
      api.buildbucket.try_build(project='infra/infra_internal'))

  superproject_change = GerritChange(
      host='chromium-review.googlesource.com',
      project='infra/infra_superproject',
      change=456789,
      patchset=12,
  )
  yield (api.test('basic-superproject') + api.buildbucket.try_build(
      gerrit_changes=[superproject_change], project='infra'))
  yield (api.test('basic-superproject-internal') + api.buildbucket.try_build(
      gerrit_changes=[superproject_change], project='infra-internal'))

  yield (
      api.test('basic-luci-go') +
      api.buildbucket.try_build(project='infra/luci/luci-go'))
