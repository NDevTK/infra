# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import contextlib
import json

from recipe_engine.recipe_api import Property

DEPS = [
  'depot_tools/bot_update',
  'depot_tools/cipd',
  'depot_tools/depot_tools',
  'depot_tools/gclient',
  'depot_tools/infra_paths',
  'recipe_engine/buildbucket',
  'recipe_engine/context',
  'recipe_engine/file',
  'recipe_engine/json',
  'recipe_engine/path',
  'recipe_engine/platform',
  'recipe_engine/properties',
  'recipe_engine/python',
  'recipe_engine/runtime',
  'recipe_engine/step',

  'infra_system',
  'infra_cipd',
]


# Mapping from a builder name to a list of GOOS-GOARCH variants it should build
# CIPD packages for. 'native' means "do not cross-compile, build for the host
# platform". Targeting 'native' will also usually build non-go based packages.
#
# If the builder is not in this set, or the list of GOOS-GOARCH for it is empty,
# it won't be used for building CIPD packages.
CIPD_PACKAGE_BUILDERS = {
  # trusty-32 is the primary builder for linux-386.
  'infra-continuous-precise-32': [],
  'infra-continuous-trusty-32':  ['native'],

  # trusty-64 is the primary builder for linux-amd64, and the rest just
  # cross-compile to different platforms (to speed up the overall cycle time by
  # doing stuff in parallel).
  'infra-continuous-precise-64': ['linux-arm', 'linux-arm64'],
  'infra-continuous-trusty-64':  ['native'],
  'infra-continuous-xenial-64':  ['linux-mips64'],
  'infra-continuous-yakkety-64': ['linux-s390x'],
  'infra-continuous-zesty-64':   ['linux-ppc64', 'linux-ppc64le'],

  # 10.13 is the primary builder for darwin-amd64.
  'infra-continuous-mac-10.10-64': [],
  'infra-continuous-mac-10.11-64': [],
  'infra-continuous-mac-10.12-64': [],
  'infra-continuous-mac-10.13-64': ['native'],

  # Windows builders each build and test for their own bitness.
  'infra-continuous-win-32': ['native'],
  'infra-continuous-win-64': ['native'],

  # Internal builders, they use exact same recipe.
  'infra-internal-continuous-trusty-64': ['native', 'linux-arm', 'linux-arm64'],
  'infra-internal-continuous-trusty-32': ['native'],
  'infra-internal-continuous-win-32': ['native'],
  'infra-internal-continuous-win-64': ['native'],
  'infra-internal-continuous-mac-10.10-64': [],
  'infra-internal-continuous-mac-10.11-64': [],
  'infra-internal-continuous-mac-10.13-64': ['native'],


  # Builders also upload CIPD packages.
  'infra-packager-linux-32': ['native'],
  'infra-packager-linux-64': [
    'native',
    'linux-arm',
    'linux-arm64',
    'linux-mips64'
    'linux-ppc64',
    'linux-ppc64le',
    'linux-s390x'
  ],
  'infra-packager-mac-64': ['native'],
  'infra-packager-win-32': ['native'],
  'infra-packager-win-64': ['native'],

  'infra-internal-packager-linux-32': ['native'],
  'infra-internal-packager-linux-64': ['native', 'linux-arm', 'linux-arm64'],
  'infra-internal-packager-mac-64': ['native'],
  'infra-internal-packager-win-32': ['native'],
  'infra-internal-packager-win-64': ['native'],
}


# A builder responsible for calling "deps.py bundle" to generate cipd bundles
# with vendored go code. We need only one.
GO_DEPS_BUNDLING_BUILDER = 'infra-continuous-trusty-64'


PROPERTIES = {
  'buildername': Property(),
  'official': Property(default=False, kind=bool,
                       help='if True, uploads packaged artifacts'),
}


def RunSteps(api, buildername, official):
  if (buildername.startswith('infra-internal-continuous') or
      buildername.startswith('infra-internal-packager')):
    project_name = 'infra_internal'
    repo_url = 'https://chrome-internal.googlesource.com/infra/infra_internal'
  elif (buildername.startswith('infra-continuous') or
      buildername.startswith('infra-packager')):
    project_name = 'infra'
    repo_url = 'https://chromium.googlesource.com/infra/infra'
  else:  # pragma: no cover
    raise ValueError(
        'This recipe is not intended for builder %s. ' % buildername)

  # Prefix the system binary path to PATH so that all Python invocations will
  # use the system Python. This will ensure that packages built will be built
  # aginst the system Python's paths.
  #
  # This is needed by the "infra_python" CIPD package, which incorporates the
  # checkout's VirtualEnv into its packages. This, in turn, results in the CIPD
  # package containing a reference to the Python that was used to create it. In
  # order to control for this, we ensure that the Python is a system Python,
  # which resides at a fixed path.
  api.gclient.set_config(project_name)
  with api.infra_system.system_env():
    bot_update_step = api.bot_update.ensure_checkout()
    api.gclient.runhooks()

    # Whatever is checked out by bot_update. It is usually equal to
    # api.properties['revision'] except when the build was triggered manually
    # ('revision' property is missing in that case).
    rev = bot_update_step.presentation.properties['got_revision']
    build_main(api, buildername, official, project_name, repo_url, rev)


def build_main(api, buildername, official, project_name, repo_url, rev):
  with api.step.defer_results():
    with api.context(cwd=api.path['checkout']):
      # Run Linux tests everywhere, Windows tests only on public CI.
      if api.platform.is_linux or project_name == 'infra':
        # TODO(tandrii): maybe get back coverage on 32-bit once
        # http://crbug/766416 is resolved.
        args = ['test']
        if (api.platform.is_linux and api.platform.bits == 32 and
            project_name == 'infra_internal'):  # pragma: no cover
          args.append('--no-coverage')
        api.python('infra python tests', 'test.py', args)

      # Validate ccompute configs.
      if api.platform.is_linux and project_name == 'infra_internal':
        api.python(
            'ccompute config test',
            'ccompute/scripts/ccompute_config.py', ['test'])

    # This downloads Go third parties, so that the next step doesn't have junk
    # output in it.
    api.python(
        'go third parties',
        api.path['checkout'].join('go', 'env.py'),
        ['go', 'version'])

    # Call 'deps.py bundle' to package dependencies specified in deps.lock into
    # a CIPD package. This is not strictly necessary, but it significantly
    # reduces time it takes to run 'env.py'. Note that 'deps.py' requires
    # environment produced by 'env.py' (for things like glide and go itself).
    # When the recipe runs with outdated deps bundle, 'env.py' call above falls
    # back to fetching dependencies from git directly. When the bundle is
    # up-to-date, 'deps.py bundle' finishes right away not doing anything.
    if buildername == GO_DEPS_BUNDLING_BUILDER:
      api.python(
          'bundle go deps',
          api.path['checkout'].join('go', 'env.py'),
          [
            'python',  # env.py knows how to expand 'python' into sys.executable
            api.path['checkout'].join('go', 'deps.py'),
            'bundle',
            '--service-account-json',
            api.cipd.default_bot_service_account_credentials,
          ])

    api.python(
        'infra go tests',
        api.path['checkout'].join('go', 'env.py'),
        ['python', api.path['checkout'].join('go', 'test.py')])

  for plat in CIPD_PACKAGE_BUILDERS.get(buildername, []):
    if plat == 'native':
      goos, goarch = None, None
    else:
      goos, goarch = plat.split('-', 1)
    with api.infra_cipd.context(api.path['checkout'], goos, goarch):
      api.infra_cipd.build()
      api.infra_cipd.test(skip_if_cross_compiling=True)
      if official:
        api.infra_cipd.upload(api.infra_cipd.tags(repo_url, rev))


def GenTests(api):

  def test(name, is_luci=False, is_experimental=False, **properties):
    default = {
      'path_config': 'generic',
      'buildername': 'infra-continuous-trusty-64',
      'buildnumber': 123,
      'mastername': 'chromium.infra',
      'repository': 'https://chromium.googlesource.com/infra/infra',
    }
    for k, v in default.iteritems():
      properties.setdefault(k, v)
    return (
        api.test(name) +
        api.runtime(is_luci=is_luci, is_experimental=is_experimental) +
        api.properties.git_scheduled(**properties)
    )

  yield test(
      'infra-continuous-precise-64',
      buildername='infra-continuous-precise-64',
      official=True)
  yield test(
      'infra-continuous-trusty-64',
      buildername='infra-continuous-trusty-64',
      official=True)

  yield (
    test(
      'infra-continuous-win-64',
      buildername='infra-continuous-win-64',
      official=True) +
    api.platform.name('win'))

  yield test(
      'infra-internal-continuous',
      buildername='infra-internal-continuous-trusty-32',
      official=True,
      mastername='internal.infra',
      repository=
          'https://chrome-internal.googlesource.com/infra/infra_internal')

  for official in [True, False]:
    yield test(
        'infra-internal-continuous-luci' + ('-official' if official else ''),
        is_luci=True,
        is_experimental=True,
        buildername='infra-internal-continuous-trusty-32',
        official=official,
        repository=
            'https://chrome-internal.googlesource.com/infra/infra_internal',
        buildbucket=json.dumps({
          "build": {
            "bucket": "luci.infra-internal.ci",
            "created_by": "user:luci-scheduler@appspot.gserviceaccount.com",
            "created_ts": 1527292217677440,
            "id": "8945511751514863184",
            "project": "infra-internal",
            "tags": [
              "builder:infra-internal-continuous-trusty-32",
              ("buildset:commit/gitiles/chrome-internal.googlesource.com/" +
                "infra/infra_internal/" +
                "+/2d72510e447ab60a9728aeea2362d8be2cbd7789"),
              "gitiles_ref:refs/heads/master",
              "scheduler_invocation_id:9110941813804031728",
              "user_agent:luci-scheduler",
            ],
          },
          "hostname": "cr-buildbucket.appspot.com"
        }),
      )

  for official in [True, False]:
    yield test(
        'infra-internal-packager' + ('-official' if official else ''),
        is_luci=True,
        is_experimental=True,
        buildername='infra-internal-packager-linux-32',
        official=official,
        repository=
            'https://chrome-internal.googlesource.com/infra/infra_internal',
        buildbucket=json.dumps({
          "build": {
            "bucket": "luci.infra-internal.prod",
            "created_by": "user:luci-scheduler@appspot.gserviceaccount.com",
            "created_ts": 1527292217677440,
            "id": "8945511751514863184",
            "project": "infra-internal",
            "tags": [
              "builder:infra-internal-packager-linux-32",
              ("buildset:commit/gitiles/chrome-internal.googlesource.com/" +
                "infra/infra_internal/" +
                "+/2d72510e447ab60a9728aeea2362d8be2cbd7789"),
              "gitiles_ref:refs/heads/master",
              "scheduler_invocation_id:9110941813804031728",
              "user_agent:luci-scheduler",
            ],
          },
          "hostname": "cr-buildbucket.appspot.com"
        }),
      )
