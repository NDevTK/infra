# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Pushes a trivial CL to Gerrit to verify git authentication works on LUCI."""
import json
from recipe_engine import post_process

PYTHON_VERSION_COMPATIBILITY = 'PY2+3'

DEPS = [
    'recipe_engine/buildbucket',
    'recipe_engine/context',
    'recipe_engine/file',
    'recipe_engine/path',
    'recipe_engine/platform',
    'recipe_engine/properties',
    'recipe_engine/step',
    'recipe_engine/time',
]


PLAYGROUND_REPO = 'https://chromium.googlesource.com/playground/access_test'


def RunSteps(api):
  #TODO(crbug/1040685): remove it after testing
  api.buildbucket.hide_current_build_in_gerrit()
  api.buildbucket.hide_current_build_in_gerrit()
  tags = api.buildbucket.tags(k1='v1', k2=['v2', 'v2', 'v2_1'])
  api.buildbucket.add_tags_to_current_build(tags)

  root_dir = api.path['tmp_base'].join('repo')
  api.file.ensure_directory('make dir', root_dir)

  with api.context(cwd=root_dir):
    api.step('git clone', ['git', 'clone', PLAYGROUND_REPO, '.'])
    api.step('git checkout -b', ['git', 'checkout', '-b', 'cl'])
    api.file.write_text(
        'drop file', root_dir.join('time.txt'), str(api.time.time()))
    api.step('git add', ['git', 'add', 'time.txt'])
    api.step('git commit', ['git', 'commit', '-m', 'Test commit'])
    api.step('push for review',
             ['git', 'push', 'origin', 'HEAD:refs/for/refs/heads/main'])

  #TODO(yuanjunh@): remove it after finish testing
  # Make sure the build.output.properties is larger than 1MB.
  with api.step.nest('set output properties') as step:
    step.properties['int'] = 1
    step.properties['boo'] = True
    step.properties['small_str'] = "abc"
    step.properties['array'] = ["a", "b"]

    large_obj = {}
    large_val = '''
    largeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
    largeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
    '''
    for i in range(0, 10000):
      large_obj['prop%d' % i] = {
          'val': large_val,
      }
    step.properties['large_obj'] = json.dumps(large_obj)


def GenTests(api):
  yield (api.test('linux') + api.platform.name('linux') +
         api.properties.generic(
             buildername='test_builder', mastername='test_master') +
         api.post_check(post_process.StatusSuccess) +
         api.post_process(post_process.DropExpectation))
