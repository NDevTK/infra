# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import ast
import re
import urllib

DEPS = [
  'build/zip',
  'depot_tools/git'
  'depot_tools/git_cl'
  'recipe_engine/cipd',
  'recipe_engine/file',
]

# Get at most 3 milestone versions released into Beta
NUM_MILESTONE_RELEASES = 3

# CIPD package path.
# https://chrome-infra-packages.appspot.com/p/chromium/testing/weblayer-x86/+/
CIPD_PKG_PATH='chromium/testing/weblayer-x86'

CHROMIUM_VERSION_REGEX = r'\d+\.\d+\.\d+\.\d+$'

# Chromium dash request constants
# TODO(rmhasan): Create a chromiumdash API to send requests to chromiumdash
CHROMIUMDASH = 'https://chromiumdash.appspot.com'
FETCH_RELEASES = '/fetch_releases?'
FETCH_COMMIT = '/fetch_commit?'

# Weblayer variants.pyl identifierr templates
WEBLAYER_NTH_NODE = 'WEBLAYER_%s_SKEW_TESTS_NTH_MILESTONE'
WEBLAYER_NTH_MINUS_ONE_NODE = (
    'WEBLAYER_%s_SKEW_TESTS_NTH_MINUS_ONE_MILESTONE')
WEBLAYER_NTH_MINUS_TWO_NODE = (
    'WEBLAYER_%s_SKEW_TESTS_NTH_MINUS_TWO_MILESTONE')

# Client and Implmentation templates for skew test configurations
CLIENT = 'CLIENT'
IMPL = 'IMPL'
CLIENT_ARGS_TEMPLATE = """'args': [
    '--test-runner-outdir',
    '.',
    '--client-outdir',
    '../../weblayer_instrumentation_test_M{milestone}/out/Release',
    '--implementation-outdir',
    '.',
    '--test-expectations',
    '../../weblayer/browser/android/javatests/skew/expectations.txt',
    '--client-version={milestone}',
],
"""
IMPL_ARGS_TEMPLATE = """'args': [
    '--test-runner-outdir',
    '.',
    '--client-outdir',
    '.',
    '--implementation-outdir',
    '../../weblayer_instrumentation_test_M{milestone}/out/Release',
    '--test-expectations',
    '../../weblayer/browser/android/javatests/skew/expectations.txt',
    '--impl-version={milestone}',
],
"""
CLIENT_ID_TEMPLATE = "'identifier': 'M{milestone}_Client_Library_Tests'"
IMPL_ID_TEMPLATE = "'identifier: M{milestone}_Implementation_Library_Tests'"

# Swarming arguments templates for all skew tests
SWARMING_ARGS = """'swarming': {
  'cipd_packages': [
    {
      'cipd_package': 'chromium/testing/weblayer-x86',
      'location': 'weblayer_instrumentation_test_M{milestone}',
      'revision': 'version:{chromium_version}',
    }
  ],
},
"""


def releases_url(platform, channel, num):
  return CHROMIUMDASH + FETCH_RELEASES + urllib.urlencode(
      {'platform': platform, 'channel': channel, 'num': num})


def commit_url(hash_value):
  return CHROMIUMDASH + FETCH_COMMIT + urllib.urlencode(
      {'commit': hash_value})


def get_chromium_version(api, hash_value):
  commit_info = api.json.load(urllib.urlopen(commit_url(hash_value)))
  version =  commit_info['deployment']['beta']
  api.assertions.assertTrue(re.match(CHROMIUM_VERSION_REGEX, version))
  return version


def is_higher_version(version, query_version):
  for p1, p2 in zip(version.split('.'), query_version.split('.')):
    if int(p2) > int(p1):
      return True
  return False


def RunSteps(api):
  # Set Gclient config to android
  api.gclient.set_config('android')
  # Checkout chromium/src at ToT
  api.bot_update.ensure_checkout()
  # Set up git config
  api.git('config', 'user.name', 'Weblayer Skew Tests Version Updates',
          name='set git config user.name')
  # Configure git cl
  api.git_cl.set_config('basic')
  api.git_cl.c.repo_location = api.path['checkout']
  # Get variants.pyl path
  variants_pyl = api.path['checkout'].join(
      'testing', 'buildbot', 'variants.pyl')
  # Get AST from variants.pyl
  variants_lines = api.file.read_text(
      'Read variants.pyl', variants_pyl).splitlines()
  variants_ast = ast.literal_eval(variants_content)
  # TODO(rmhasan): Create wget API and use it instead of urllib
  # Get Chromium project commits released into the beta channel
  # Get atleast the last 20 commits released so that we can possibly
  # get 3 milestone versions below
  beta_releases = api.json.load(
      urllib.urlopen(releases_url('Android', 'Beta', 20)))
  # Convert hashes into chromium version numbers
  chromium_versions = [get_chromium_version(
                           api, beta_release['hashes']['chromium'])
                       for beta_release in beta_releases]

  # Map milestone number to milestone version numbers
  milestones_to_versions = {}
  for version in chromium_versions:
    milestone = int(version[:version.index('.')])
    milestones_to_versions.setdefault(milestone, '0.0.0.0')
    if is_higher_version(milestone_to_versions[milestone], version):
      milestones_to_versions[milestone] = version

  # Sort milestone versions by milestone number
  recent_milestone_versions = [
      milestones_to_versions[milestone] for milestone in
      sorted(milestones_to_versions, reverse=True)]

  for version, node_name in zip(
      recent_milestone_versions, [node % CLIENT for node in [
          WEBLAYER_NTH_NODE, WEBLAYER_NTH_MINUS_ONE_NODE,
          WEBLAYER_NTH_MINUS_TWO_NODE]]):
    pass

  # Prepare staging directory to unpack gsutil into.
  staging_dir = api.path['start_dir'].join('gsutil_staging_dir')
  api.file.rmtree('cleaning staging dir', staging_dir)

  try:
    version = _gsutil_version(api)
    name = 'gsutil_%s.zip' % version
    url = '%s/%s' % (GSUTIL_BUCKET, name)
    gsutil_zip = api.path['start_dir'].join(name)

    api.gsutil.download_url(url, gsutil_zip, name='Download %s' % name)
    api.zip.unzip('Unzip %s' % name, gsutil_zip, staging_dir, quiet=True)

    gsutil_dir = staging_dir.join('gsutil')
    api.path.mock_add_paths(gsutil_dir)
    assert api.path.exists(gsutil_dir), (
        'Package directory %s does not exist' % (gsutil_dir))

    # Build and register our CIPD package.
    api.cipd.build(
        input_dir=gsutil_dir,
        output_package=cipd_pkg_dir,
        package_name=cipd_pkg_name,
    )
    api.cipd.register(
        package_name=cipd_pkg_name,
        package_path=cipd_pkg_dir,
        refs=['latest'],
        tags={'gsutil_version': version},
    )
  finally:
    api.file.remove('remove gsutil directory', cipd_pkg_dir)


def GenTests(api):
  yield (
    api.test('linux') +
    api.platform.name('linux')
  )
