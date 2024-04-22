# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PYTHON_VERSION_COMPATIBILITY = "PY2+3"

DEPS = [
    'recipe_engine/context',
    'recipe_engine/file',
    'recipe_engine/path',
    'recipe_engine/platform',
    'recipe_engine/step',
    'zip',
]


def RunSteps(api):
  # Prepare files.
  temp = api.path.mkdtemp('zip-example')
  api.step('touch a', ['touch', temp / 'a'])
  api.step('touch b', ['touch', temp / 'b'])
  api.file.ensure_directory('mkdirs', temp.joinpath('sub', 'dir'))
  api.step('touch c', ['touch', temp.joinpath('sub', 'dir', 'c')])

  # Build zip using 'zip.directory'.
  api.zip.directory('zipping', temp, temp / 'output.zip', comment='hello')

  # Build a zip using ZipPackage api.
  package = api.zip.make_package(temp, temp / 'more.zip')
  package.add_file(package.root / 'a')
  package.add_file(package.root / 'b')
  package.add_directory(package.root / 'sub')
  package.zip('zipping more')

  # Update a zip using ZipPackage api.
  package = api.zip.update_package(temp, temp / 'more.zip')
  package.add_file(temp / 'update_a', 'renamed_a')
  package.add_file(temp / 'update_b', 'renamed_b')
  package.set_comment('hello again')
  package.zip('zipping more updates')

  # Coverage for 'output' property.
  api.step('report', ['echo', package.output])

  # Unzip the package.
  api.zip.unzip(
      'unzipping',
      temp.joinpath('output.zip'),
      temp.joinpath('output'),
      quiet=True)
  # List unzipped content.
  with api.context(cwd=temp / 'output'):
    api.step('listing', ['find'])
  # Clean up.
  api.file.rmtree('cleanup', temp)

  # Retrieve archive comment.
  comment = api.zip.get_comment('get comment', temp / 'output.zip')
  api.step('report comment', ['echo', comment])


def GenTests(api):
  for platform in ('linux', 'win', 'mac'):
    yield api.test(
        platform,
        api.platform.name(platform),
    )
