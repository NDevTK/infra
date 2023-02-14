# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import textwrap

from recipe_engine.recipe_api import Property

PYTHON_VERSION_COMPATIBILITY = "PY2+3"

DEPS = [
    'depot_tools/depot_tools',
    'depot_tools/osx_sdk',
    'infra_checkout',
    'infra_system',
    'recipe_engine/buildbucket',
    'recipe_engine/context',
    'recipe_engine/file',
    'recipe_engine/path',
    'recipe_engine/platform',
    'recipe_engine/properties',
    'recipe_engine/raw_io',
    'recipe_engine/resultdb',
    'recipe_engine/runtime',
    'recipe_engine/step',
]


PROPERTIES = {
    'go_version_variant':
        Property(
            default=None,
            kind=str,
            help='A go version variant to bootstrap, see bootstrap.py'),
    'run_lint':
        Property(default=False, kind=bool, help='Whether to run linter'),
}


def RunSteps(api, go_version_variant, run_lint):
  cl = api.buildbucket.build.input.gerrit_changes[0]
  project = cl.project
  assert project in ('infra/infra', 'infra/infra_internal'), (
      'unknown project: "%s"' % project)
  patch_root = project.split('/')[-1]
  internal = (patch_root == 'infra_internal')

  co = api.infra_checkout.checkout(
      gclient_config_name=patch_root, patch_root=patch_root,
      internal=internal,
      generate_env_with_system_python=True,
      go_version_variant=go_version_variant)
  co.commit_change()
  co.gclient_runhooks()

  if run_lint:
    with api.context(
        cwd=co.path.join('infra', 'go', 'src', 'infra')), co.go_env():
      api.infra_checkout.apply_golangci_lint(co, 'go/src/infra/')
    return

  # Analyze the CL to skip unnecessary tests.
  files = co.get_changed_files()
  is_deps_roll = bool(files.intersection([
      'DEPS',
      api.path.join('bootstrap', 'deps.pyl'),
      api.path.join('go', 'deps.lock')
  ]))
  is_build_change = any(f.startswith('build/') for f in files)
  is_go_change = any(f.startswith('go/') for f in files)
  is_pure_go_change = all(f.startswith('go/') for f in files)

  # Don't run Python or recipes tests if only "go/..." was touched.
  if not is_pure_go_change:
    with api.step.defer_results():
      if api.platform.arch != 'arm':
        with api.context(cwd=co.path.join(patch_root)):
          api.step('python tests', ['python3', 'test.py', 'test', '--verbose'])
        if internal and (api.platform.is_linux or api.platform.is_mac) and any(
            f.startswith('appengine/chromiumdash') for f in files):
          cwd = api.path['checkout'].join('appengine', 'chromiumdash')
          gae_env = {
              'GAE_RUNTIME': 'python3',
              'GAE_APPLICATION': 'testbed-test',
          }
          with api.context(cwd=cwd, env=gae_env):
            api.step('chromiumdash python3 tests', [
                'vpython3', '-m', 'pytest', '--ignore=gae_ts_mon/',
                '--ignore=go/'
            ])

      if not internal and api.platform.is_linux and api.platform.bits == 64:
        api.step('recipe test', [
            'python3',
            co.path.join('infra', 'recipes', 'recipes.py'), 'test', 'run'
        ])
        api.step(
            'recipe lint',
            ['python3',
             co.path.join('infra', 'recipes', 'recipes.py'), 'lint'])
  else:
    api.step('skipping Python tests for pure Go change', cmd=None)

  # Don't run Go tests unless go/... or DEPS or build/ were touched.
  if not (is_deps_roll or is_build_change or is_go_change):
    api.step('skipping Go and CIPD packaging tests', cmd=None)
    return

  # Some third_party go packages on OSX rely on cgo and thus a configured
  # clang toolchain.
  with api.osx_sdk('mac'), co.go_env():
    with api.depot_tools.on_path():
      # Some go tests test interactions with depot_tools binaries, so put
      # depot_tools on the path.
      api.step(
          'go tests',
          api.resultdb.wrap(
              ['vpython3', '-u',
               co.path.join(patch_root, 'go', 'test.py')]))


    # Do slow *.cipd packaging tests only when touching build/* or DEPS. This
    # will build all registered packages (without uploading them), and run
    # package tests from build/tests/.
    if is_build_change or is_deps_roll:
      api.step(
          'cipd - build packages',
          ['vpython', co.path.join(patch_root, 'build', 'build.py')])
      api.step(
          'cipd - test packages integrity',
          ['vpython',
           co.path.join(patch_root, 'build', 'test_packages.py')])
      if api.platform.is_win:
        with api.context(env={'GOOS': 'windows', 'GOARCH': 'arm64'}):
          api.step('cipd - build packages (ARM64)',
                   ['vpython',
                    co.path.join(patch_root, 'build', 'build.py')])
          # Cross-compiling, so no tests.
    else:
      api.step('skipping slow CIPD packaging tests', cmd=None)


def GenTests(api):
  def diff(*files):
    return api.step_data('get change list',
                         api.raw_io.stream_output_text('\n'.join(files)))

  def test(name, internal=False, buildername='generic tester'):
    return (
        api.test(name) + api.runtime(is_experimental=True) +
        api.buildbucket.try_build(
            project='infra-internal' if internal else 'infra',
            builder=buildername,
            git_repo=(
                'https://chrome-internal.googlesource.com/infra/infra_internal'
                if internal else
                'https://chromium.googlesorce.com/infra/infra')))

  yield (
    test('basic') +
    diff('infra/stuff.py', 'go/src/infra/stuff.go')
  )

  yield (
    test('basic_arm64') +
    api.platform.arch('arm') +
    diff('infra/stuff.py', 'go/src/infra/stuff.go')
  )

  yield (
    test('basic_internal', internal=True) +
    diff('infra/stuff.py', 'go/src/infra/stuff.go')
  )

  yield (
    test('only_go') +
    diff('go/src/infra/stuff.go')
  )

  yield (
    test('only_go_override_version') +
    api.properties(go_version_variant='bleeding_edge') +
    diff('go/src/infra/stuff.go')
  )

  yield (
    test('only_go_osx') +
    api.platform('mac', 64) +
    diff('go/src/infra/stuff.go')
  )

  yield (
    test('only_js') +
    diff('appengine/foo/static/stuff.js')
  )

  yield (test('infra_internal_with_chromium_dash', internal=True) +
         diff('appengine/chromiumdash/foo.py'))

  yield (
    test('only_cipd_build') +
    diff('build/build.py')
  )

  yield (test('only_cipd_build_win') + api.platform('win', 64) +
         diff('build/build.py'))

  yield (test('lint_try_job') + api.properties(run_lint=True) + api.step_data(
      'get change list',
      stdout=api.raw_io.output_text(
          textwrap.dedent("""\
      go/src/infra/cmd/tools.go
      """))))
