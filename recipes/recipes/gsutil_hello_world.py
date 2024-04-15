# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Pushes a trivial CL to Gerrit to verify git authentication works on LUCI."""

PYTHON_VERSION_COMPATIBILITY = 'PY2+3'

DEPS = [
  'depot_tools/depot_tools',
  'depot_tools/gsutil',
  'recipe_engine/file',
  'recipe_engine/path',
  'recipe_engine/platform',
  'recipe_engine/step',
  'recipe_engine/time',
]


def RunSteps(api):
  root_dir = api.path.tmp_base_dir
  name = 'access_test'

  api.file.write_text('write %s' % name, root_dir.join(name),
                      str(api.time.time()))
  api.gsutil.upload(root_dir.join(name), 'luci-playground', name)
  api.step(
      'upload_to_google_storage',
      ['python3', api.depot_tools.upload_to_google_storage_path,
       '-b', 'luci-playground', root_dir.join(name)])


def GenTests(api):
  yield api.test('linux') + api.platform.name('linux')
