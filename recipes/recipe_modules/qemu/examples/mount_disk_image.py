# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import DropExpectation, StatusSuccess
from PB.recipes.infra.windows_image_builder import input as input_pb

DEPS = [
    'qemu', 'recipe_engine/path', 'recipe_engine/properties',
    'recipe_engine/raw_io'
]

PYTHON_VERSION_COMPATIBILITY = 'PY3'

PROPERTIES = input_pb.Inputs


# the recipe that we are testing
def RunSteps(api, inputs):
  api.qemu.init('latest')
  if inputs.config_path.endswith('.iso'):
    # set partitions to None for iso images
    api.qemu.mount_disk_image(inputs.config_path, partitions=None)
  else:
    # For others just use the first partition
    api.qemu.mount_disk_image(inputs.config_path)


# the tests that we are doing on the recipe
def GenTests(api):
  # test mounting an iso cdrom image
  yield (api.test('Test mount iso') +
         api.properties(input_pb.Inputs(config_path='test/windows_10.iso')) +
         api.step_data(
             'Mount loop',
             api.raw_io.stream_output(
                 'Mounted /dev/loop6 at /media/chrome-bot/test'),
             retcode=0) + api.post_process(StatusSuccess) +
         api.post_process(DropExpectation))

  # test mounting the first partition of a disk image
  yield (api.test('Test mount image') +
         api.properties(input_pb.Inputs(config_path='test/system.img')) +
         api.step_data(
             'Mount loop',
             api.raw_io.stream_output(
                 'Mounted /dev/loop6 at /media/chrome-bot/test'),
             retcode=0) + api.post_process(StatusSuccess) +
         api.post_process(DropExpectation))
