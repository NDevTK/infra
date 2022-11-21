# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

DEPS = ['recipe_engine/step',
        'recipe_engine/cipd',
        'recipe_engine/path']

def RunSteps(api):
  api.step('Print Hello World', ['echo', 'hello', 'world'])
  ef = api.cipd.EnsureFile()
  ef.add_package(name='experimental/jairogarciga_at_google.com/purple_panda',
                 version='latest')
  api.cipd.ensure(root=api.path['cache'], ensure_file=ef)
  api.step('Check what we have', ['ls', api.path['cache']])

def GenTests(api):
 yield api.test('basic')
