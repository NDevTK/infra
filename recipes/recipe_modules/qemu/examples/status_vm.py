# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import DropExpectation, StatusSuccess
from recipe_engine.post_process import DropExpectation, StatusException
from recipe_engine.recipe_api import Property

DEPS = [
    'qemu', 'recipe_engine/raw_io', 'recipe_engine/json', 'recipe_engine/path'
]

PYTHON_VERSION_COMPATIBILITY = 'PY3'


def RunSteps(api):
  api.qemu.vm_status(name='test')


def GenTests(api):

  # Test vm_status successful path
  yield (api.test('Test status') + api.post_process(StatusSuccess) +
         api.step_data(
             'Status test',
             stdout=api.json.output({
                 'return': {
                     'running': True,
                     'singlestep': False,
                     'status': 'running'
                 }
             })) + api.post_process(DropExpectation))

  # Test vm_status fail, JSON returned by QMP
  yield (api.test('Test status fail') + api.post_process(StatusSuccess) +
         api.step_data('Status test', stdout=api.json.output({
             'return': {},
         })) + api.post_process(DropExpectation))

  # Test vm_stanus fail, failed to connect to socket, no exception thrown
  yield (api.test('Test status missing socket') +
         api.post_process(StatusSuccess) + api.step_data(
             'Status test',
             stdout=api.json.output(
                 {'return': {
                     'Error': '[Errno 111] Connection refused'
                 }}),
             retcode=0) + api.post_process(DropExpectation))

  # Test vm_status fail, failed to find the python script to call. (should not
  # happen)
  yield (api.test('Test status missing python script') +
         api.post_process(StatusException) + api.step_data(
             'Status test',
             stdout=api.raw_io.output("""
                /bin/bash: line 1: something/qemu/resources/qmp.py: No such file or directory
            """),
             retcode=127) + api.expect_exception('QEMUError') +
         api.post_process(DropExpectation))
