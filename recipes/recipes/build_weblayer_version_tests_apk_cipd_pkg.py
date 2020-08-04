# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import ast

DEPS = [
  'build/zip',
  'recipe_engine/cipd',
  'recipe_engine/file',
  'recipe_engine/path',
  'recipe_engine/platform',
  'recipe_engine/properties',
  'recipe_engine/python',
  'recipe_engine/raw_io',
  'recipe_engine/step',
]

# CIPD package path.
# https://chrome-infra-packages.appspot.com/p/chromium/testing/weblayer-x86/+/
CIPD_PKG_PATH='chromium/testing/weblayer-x86'

CHROMIUM_VERSION_REGEX = r'\d+\.\d+\.\d+\.\d+$'


def RunSteps(api):
  # Set Gclient config to android
  api.gclient.set_config('android')
  # Checkout most current src/ code
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
  variants_content = api.file.read_text('Read variants.pyl', variants_pyl)
  variants_ast = ast.literal_eval(variants_content)
  

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
