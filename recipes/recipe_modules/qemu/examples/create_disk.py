# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import DropExpectation, StatusSuccess
from recipe_engine.post_process import StatusFailure
from recipe_engine.recipe_api import Property
from textwrap import dedent

DEPS = ['qemu', 'recipe_engine/path', 'recipe_engine/raw_io']

PYTHON_VERSION_COMPATIBILITY = 'PY3'


def RunSteps(api):
  api.qemu.init('latest')
  # test good cases for both create_empty_disk and create_disk
  api.qemu.create_disk('fat_disk', 'fat', 2048)
  # mock existence of cache directory
  api.path.mock_add_paths(api.path['cache'], 'DIRECTORY')
  # mock cleanup to be a file
  api.path.mock_add_paths(api.path['cleanup'], 'FILE')
  api.qemu.create_disk('ext4_disk', 'ext4', 2048,
                       {api.path['cache']: 'got_cache/i_need_it'})


def GenTests(api):
  yield (api.test('Test create_disk pass') + api.post_process(StatusSuccess) +
         api.step_data(
             'Copy files to ext4_disk.Mount loop',
             api.raw_io.stream_output(
                 'Mounted /dev/loop6 at /media/chrome-bot/test'),
             retcode=0) +
         # mock the free disk size to say there is enough
         api.step_data(
             'Check free space on disk for fat_disk',
             api.raw_io.stream_output(
                 dedent('''Avail
                           27815012
                        ''')),
             retcode=0) +
         # mock the free disk size to say there is enough
         api.step_data(
             'Check free space on disk for ext4_disk',
             api.raw_io.stream_output(
                 dedent('''Avail
                           13907506
                        ''')),
             retcode=0) + api.post_process(DropExpectation))

  yield (api.test('Test create_disk fail (mount permission)') +
         api.post_process(StatusFailure) + api.step_data(
             'Copy files to ext4_disk.Mount loop',
             api.raw_io.stream_output('Permission denied: /dev/loop6'),
             retcode=1) +
         # mock the free disk size to say there is enough
         api.step_data(
             'Check free space on disk for fat_disk',
             api.raw_io.stream_output(
                 dedent('''Avail
                           27815012
                        ''')),
             retcode=0) +
         # mock the free disk size to say there is enough
         api.step_data(
             'Check free space on disk for ext4_disk',
             api.raw_io.stream_output(
                 dedent('''Avail
                           13907506
                        ''')),
             retcode=0) + api.step_data(
                 'Copy files to ext4_disk.Mount loop',
                 api.raw_io.stream_output('Permission denied: /dev/loop6'),
                 retcode=1) + api.post_process(DropExpectation))

  yield (api.test('Test create_disk fail (out of disk)') +
         api.post_process(StatusFailure) +
         # mock the free disk size to say there is enough
         api.step_data(
             'Check free space on disk for fat_disk',
             api.raw_io.stream_output(
                 dedent('''Avail
                           27815012
                        ''')),
             retcode=0) +
         # mock the free disk size to say not enough
         api.step_data(
             'Check free space on disk for ext4_disk',
             api.raw_io.stream_output(
                 dedent('''Avail
                           13907504
                        ''')),
             retcode=0) + api.post_process(DropExpectation))
