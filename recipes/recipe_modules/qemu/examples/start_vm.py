# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import DropExpectation, StatusSuccess
from recipe_engine.post_process import StatusFailure, StepCommandRE
from recipe_engine.recipe_api import Property

DEPS = ['qemu', 'recipe_engine/raw_io', 'recipe_engine/path']

PYTHON_VERSION_COMPATIBILITY = 'PY3'


def RunSteps(api):
  api.qemu.init('latest')
  api.qemu.start_vm(name='test', arch='aarch64', memory=8192, disks=['test'])


def GenTests(api):

  yield (
      api.test('Test start vm') + api.post_process(StatusSuccess) +
      api.step_data('Start vm test',
                    api.raw_io.output('VNC server running on 127.0.0.1:5900')) +
      # check if the command was run correctly
      api.post_process(StepCommandRE, 'Start vm test', [
          '.*qemu-system-aarch64', '-qmp', 'tcp:localhost:4444,server,nowait',
          '-daemonize', '-m', '8192', '-drive',
          'file=\[CLEANUP\]/qemu/workdir/disks/test,format=raw,if=ide,'
          'media=disk,index=0'
      ]) + api.post_process(DropExpectation))

  yield (
      api.test('Test start vm fail') + api.post_process(StatusFailure) +
      api.step_data(
          'Start vm test', api.raw_io.output('Failed to start vm'), retcode=1) +
      api.post_process(DropExpectation))
