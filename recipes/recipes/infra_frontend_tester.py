# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PYTHON_VERSION_COMPATIBILITY = "PY2+3"

DEPS = [
    'recipe_engine/buildbucket',
    'recipe_engine/cipd',
    'recipe_engine/context',
    'recipe_engine/path',
    'recipe_engine/platform',
    'recipe_engine/properties',
    'recipe_engine/step',
    'recipe_engine/file',
    'depot_tools/bot_update',
    'depot_tools/gclient',
]

def RunSteps(api):
  assert api.platform.is_linux, 'Unsupported platform, only Linux is supported.'
  cl = api.buildbucket.build.input.gerrit_changes[0]
  project_name = cl.project
  assert project_name in ('infra/infra', 'infra/infra_internal',
                          'infra/luci/luci-go'), ('unknown project: "%s"' %
                                                  project_name)
  patch_root = project_name.split('/')[-1]
  config_name = patch_root.replace("-", "_")

  path = api.path['cache'].join('builder')
  api.file.ensure_directory('ensure builder dir', path)

  with api.context(cwd=path):
    api.gclient.set_config(config_name)

    api.bot_update.ensure_checkout(patch_root=patch_root)
    api.gclient.runhooks()

  packages_dir = api.path['start_dir'].join('packages')
  ensure_file = api.cipd.EnsureFile()
  ensure_file.add_package('infra/3pp/tools/nodejs/${platform}',
                          'version:2@16.13.0')
  api.cipd.ensure(packages_dir, ensure_file)

  node_path = api.path['start_dir'].join('packages', 'bin')
  env = {
      'PATH': api.path.pathsep.join([str(node_path), '%(PATH)s'])
  }
  if patch_root == 'infra':
    RunInfraFrontendTests(api, env, api.path['checkout'].join('infra'))
  elif patch_root == 'infra_internal':
    RunInfraInternalFrontendTests(api, env,
                                  api.path['checkout'].join('infra_internal'))
  else:
    RunLuciGoTests(api, env)


def RunInfraInternalFrontendTests(api, env, root_path):
  """This function runs UI tests in `infra_internal` project.
  """

  # Add your infra_internal tests here following this example:
  # cwd = api.path['checkout'].join('path', 'to', 'ui', 'root')
  # RunFrontendTests(api, env, cwd, 'myapp')
  # `myapp` is the name that will show up in the step.

  testhaus = root_path.join('go', 'src', 'infra_internal', 'appengine',
                            'testhaus')
  RunFrontendTests(api, env, testhaus.join('frontend', 'ui'), 'testhaus')

  cwd = root_path.join('go', 'src', 'infra_internal', 'appengine', 'spike',
                       'appengine', 'frontend', 'ui')
  RunFrontendTests(api, env, cwd, 'spike')


def RunInfraFrontendTests(api, env, root_path):
  """This function runs the UI tests in `infra` project.
  """

  cwd = root_path.join('appengine', 'monorail')
  RunFrontendTests(api, env, cwd, 'monorail')

  cwd = root_path.join('go', 'src', 'infra', 'appengine', 'dashboard',
                       'frontend')
  RunFrontendTests(api, env, cwd, 'chopsdash')


def RunLuciGoTests(api, env):
  """This function runs UI tests in the `luci-go` project.
  """
  # This variable defnies the base directory for luci-go project under infra
  luci_go_directory = 'go/src/go.chromium.org/luci'

  cwd = api.path['checkout'].join(luci_go_directory, 'analysis', 'frontend',
                                  'ui')
  RunFrontendTests(api, env, cwd, 'analysis')

  cwd = api.path['checkout'].join(luci_go_directory, 'bisection', 'frontend',
                                  'ui')
  RunFrontendTests(api, env, cwd, 'bisection')

  cwd = api.path['checkout'].join(luci_go_directory, 'milo', 'ui')
  RunFrontendTests(api, env, cwd, 'milo')


def RunFrontendTests(api, env, cwd, app_name):
  with api.context(env=env, cwd=cwd):
    api.step(('%s npm install' % app_name), ['npm', 'ci'])
    api.step(('%s test' % app_name), ['npm', 'run', 'test'])


def GenTests(api):
  yield (
      api.test('basic') +
      api.buildbucket.try_build(project='infra/infra'))
  yield (
      api.test('basic-internal') +
      api.buildbucket.try_build(project='infra/infra_internal'))
  yield (api.test('basic-luci-go') +
         api.buildbucket.try_build(project='infra/luci/luci-go'))
