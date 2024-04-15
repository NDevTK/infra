# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Recipe to build windows depot_tools bootstrap zipfile."""

PYTHON_VERSION_COMPATIBILITY = "PY2+3"

DEPS = [
    'depot_tools/git',
    'depot_tools/gsutil',
    'recipe_engine/buildbucket',
    'recipe_engine/file',
    'recipe_engine/json',
    'recipe_engine/path',
    'recipe_engine/platform',
    'recipe_engine/properties',
    'recipe_engine/raw_io',
    'recipe_engine/step',
    'zip',
]


REPO_URL='https://chromium.googlesource.com/chromium/tools/depot_tools.git'
DOC_UPLOAD_URL='gs://chrome-infra-docs/flat/depot_tools/docs/'

def RunSteps(api):
  # prepare the output dir and zip paths
  api.path.checkout_dir = api.path.start_dir.join('depot_tools')
  zip_out = api.path.start_dir.join('depot_tools.zip')

  api.step('mkdir depot_tools', ['mkdir', api.path.checkout_dir])

  with api.step.nest('clone + checkout'):
    api.git('clone', '-n', REPO_URL, api.path.checkout_dir)
    api.step.active_result.presentation.properties['got_revision'] = (
        api.buildbucket.gitiles_commit.id)
    api.git('config', 'core.autocrlf', 'false', name='set autocrlf')
    api.git('config', 'core.filemode', 'false', name='set filemode')
    api.git('config', 'core.symlinks', 'false', name='set symlinks')
    api.git('checkout', 'origin/main')
    api.git('reset', '--hard', api.buildbucket.gitiles_commit.id)
    api.git('reflog', 'expire', '--all')
    api.git('gc', '--aggressive', '--prune=all')

  # zip + upload repo
  api.zip.directory('zip it up', api.path.checkout_dir, zip_out)
  api.gsutil.upload(zip_out, 'chrome-infra', 'depot_tools.zip',
                    args=['-a', 'public-read'], unauthenticated_url=True)

  # upload html docs
  api.gsutil(['cp', '-r', '-z', 'html', '-a', 'public-read',
              api.path.checkout_dir.join('man', 'html'), DOC_UPLOAD_URL],
             name='upload docs')


def GenTests(api):
  yield (
      api.test('basic') +
      api.buildbucket.ci_build(git_repo=REPO_URL, revision='deadbeef')
  )
