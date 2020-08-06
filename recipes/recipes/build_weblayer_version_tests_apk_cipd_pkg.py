# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import ast
import contextlib
import os
import re
import shutil
import tempfile
import urllib

DEPS = [
  'build/chromium',
  'depot_tools/git',
  'depot_tools/git_cl',
  'recipe_engine/cipd',
  'recipe_engine/file',
  'recipe_engine/python',
]

# CIPD package path.
# https://chrome-infra-packages.appspot.com/p/chromium/testing/weblayer-x86/+/
CIPD_PKG_NAME='chromium/testing/weblayer-x86'

CHROMIUM_VERSION_REGEX = r'\d+\.\d+\.\d+\.\d+$'

# Chromium dash request constants
# TODO(rmhasan): Create a chromiumdash API to send requests to chromiumdash
CHROMIUMDASH = 'https://chromiumdash.appspot.com'
FETCH_RELEASES = '/fetch_releases?'
FETCH_COMMIT = '/fetch_commit?'

# Weblayer variants.pyl identifierr templates
WEBLAYER_NTH_TMPL = 'WEBLAYER_%s_SKEW_TESTS_NTH_MILESTONE'
WEBLAYER_NTH_MINUS_ONE_TMPL = (
    'WEBLAYER_%s_SKEW_TESTS_NTH_MINUS_ONE_MILESTONE')
WEBLAYER_NTH_MINUS_TWO_TMPL = (
    'WEBLAYER_%s_SKEW_TESTS_NTH_MINUS_TWO_MILESTONE')

# Client and Implementation templates for skew test configurations
CLIENT = 'CLIENT'
IMPL = 'IMPL'
CLIENT_ARGS_TMPL = """'args': [
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
IMPL_ARGS_TMPL = """'args': [
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
CLIENT_ID_TMPL = "'identifier': 'M{milestone}_Client_Library_Tests'"
IMPL_ID_TMPL = "'identifier: M{milestone}_Implementation_Library_Tests'"

# Swarming arguments templates for all skew tests
SWARMING_ARGS_TMPL = """'swarming': {
  'cipd_packages': [
    {
      'cipd_package': 'chromium/testing/weblayer-x86',
      'location': 'weblayer_instrumentation_test_M{milestone}',
      'revision': 'version:{chromium_version}',
    }
  ],
},
"""

def generate_skew_test_config_lines(library, version):
  lines = []
  milestone = version[:version.index('.')]
  args_tmpl = CLIENT_ARGS_TMPL if library == CLIENT else IMPL_ARGS_TMPL
  id_tmpl = CLIENT_ID_TMPL if library == CLIENT else IMPL_ID_TMPL
  lines.extend(
      ' ' * 4 + v.rstrip()
      for v in  args_tmpl.format(milestone=milestone).splitlines())
  lines.append(' ' * 4 + id_tmpl.format(milestone=milestone))
  lines.extend(
      ' ' * 4 + v.rstrip()
      for v in SWARMING_ARGS_TMPL.format(
          milestone=milestone, chromium_version=version).splitlines())
  return lines


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


@contextlib.contextmanager
def checkout_chromium_version(api, version):
  try:
    api.git('checkout', version)
    yield
  finally:
    api.git('checkout', 'origin/master')


@contextlib.contextmanager
def get_temporary_directory():
  tempdir = tempfile.mkdtemp()
  try:
    yield tempdir
  finally:
    shutil.rmtree(tempdir)


def RunSteps(api):
  # Set Gclient config to android
  api.gclient.set_config('android')
  # Checkout chromium/src at ToT
  api.bot_update.ensure_checkout()
  # Ensure that goma is installed
  api.chromium.ensure_goma()
  # Set up git config
  api.git('config', 'user.name', 'Weblayer Skew Tests Version Updates',
          name='set git config user.name')
  # Configure git cl
  api.git_cl.set_config('basic')
  api.git_cl.c.repo_location = api.path['checkout']
  # Checkout origin.master
  # Also's cd's into src/
  api.git('checkout', 'origin/master')

  # Read variants.pyl
  variants_pyl_path = api.path['checkout'].join(
      'testing', 'buildbot', 'variants.pyl')
  variants_lines = api.file.read_text(
      'Read variants.pyl', variants_pyl_path).splitlines()

  # TODO(rmhasan): Create wget API and use it instead of urllib
  # Get Chromium project commits released into the beta channel
  # Get at least the last 20 commits released so that we can possibly
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

  # Map Chromium versions to variants.pyl identifiers
  versions_to_variants_id_tmpl = zip(
      recent_milestone_versions,
      [WEBLAYER_NTH_TMPL, WEBLAYER_NTH_MINUS_ONE_TMPL,
       WEBLAYER_NTH_MINUS_TWO_TMPL])

  # TODO(crbug.com/1041619): Add presubmit check for variants.pyl
  # that checks if variants.pyl follows a format which allows the code
  # below to overwrite skew test configurations
  new_variants_lines = []
  cipd_pkgs_to_create = []
  lineno = 0
  while lineno < len(variants_lines):
    for version, tmpl in versions_to_variants_id_tmpl:
      for library in [CLIENT, IMPL]:
        variants_id = tmpl % library
        if variants_id in variants_lines[lineno]:
          new_variants_lines.append(variants_lines[lineno])
          new_variants_lines.extend(
              generate_skew_test_config_lines(library, version))
          contains_current_version = False
          while not all(c in '}, ' for c in variants_lines[lineno]):
            contains_current_version |= version in variants_lines[lineno]
            lineno += 1
          if not contains_current_version:
            cipd_pkgs_to_create.append(version)
    new_variants_lines.append(variants_lines[lineno])
    lineno += 1

  # Create CIPD packages for all chromium versions added to variants.pyl
  for version in cipd_pkgs_to_create:
    # Checkout chromium version
    with checkout_chromium_version(api, version), \
        tempfile.TemporaryFile(suffix='.zip') as zip_path, \
        get_temporary_directory() as extract_path:
      # Generate build files for weblayer instrumentation tests APK - x86
      mb_py_path = api.path['checkout'].join('tools', 'mb', 'mb.py')
      mb_config_path = api.path['checkout'].join(
          'weblayer', 'browser', 'android',
          'javatests', 'skew', 'mb_config.pyl')
      mb_args = [
         'zip', '--master=dummy.master', '--builder=dummy.builder',
         '--config-file=%s' % mb_config_path, os.path.join('out', 'Release'),
         'weblayer_instrumentation_test_apk', zip_path]
      api.python('Generating build files for weblayer_instrumentation_test_apk',
                 mb_py_path, mb_args)
      # Build weblayer instrumentation tests APK - x86 CIPD package
      with zipfile.ZipFile(zip_path) as zip_file:
        zip_file.extractall(path=extract_path)
      api.cipd.build(extract_path, 'weblayer_instrumentation_tests_apk.cipd',
                     CIPD_PKG_NAME)
      api.cipd.register(CIPD_PKG_NAME,
                        os.path.join(
                            extract_path,
                            'weblayer_instrumentation_tests_apk.cipd'),
                        tags={'version', version},
                        refs=['m%s' % version[:version.index('.')]])

  if cipd_pkgs_to_create:
    # New chromium versions were added to variants.pyl so we need to write
    # new changes to src/testing/buildbot/variants.pyl and commit them to the
    # main branch
    api.file.write_text(
        'Write to variants.py', variants_pyl_path,
        '\n'.join(new_variants_lines))


def GenTests(api):
  yield api.test('basic')
