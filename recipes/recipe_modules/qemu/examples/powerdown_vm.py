# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import DropExpectation, StatusSuccess
from recipe_engine.recipe_api import Property

DEPS = [
    'qemu', 'recipe_engine/raw_io', 'recipe_engine/json', 'recipe_engine/path'
]

PYTHON_VERSION_COMPATIBILITY = 'PY3'


def RunSteps(api):
  assert (api.qemu.powerdown_vm(name='test'))


def GenTests(api):
  # Happy path. We sent `system_powerdown` signal and got
  # {
  #   "return": {}
  # }
  yield (
      api.test('Test powerdown_vm') + api.post_process(StatusSuccess) +
      api.step_data(
          'Powerdown test', stdout=api.json.output({
              'return': {},
          }), retcode=0) + api.post_process(DropExpectation))

  # Failed as VM is already down. We sent `system_powerdown` signal and got
  # {
  #   "return": {
  #       "Error": "[Errno 111] Connection refused"
  #   }
  # }.
  # This is still a happy path. As VM is already down
  yield (api.test('Test powerdown_vm vm not running') +
         api.post_process(StatusSuccess) + api.step_data(
             'Powerdown test',
             stdout=api.json.output({
                 'return': {
                     'Error': '[Errno 111] Connection refused'
                 },
             }),
             retcode=0) + api.post_process(DropExpectation))

  # We sent `system_powerdown` signal and got
  # {
  #   "return": {
  #       "Error": "QMP FAILURE"
  #   }
  # }
  yield (api.test('Test powerdown_vm qmp failure') + api.step_data(
      'Powerdown test',
      stdout=api.json.output({
          'return': {
              'Error': 'QMP FAILURE'
          },
      }),
      retcode=0) + api.expect_exception('AssertionError') +
         api.post_process(DropExpectation))

  # Test failure to powerdown_vm. Failure to find qmp.py
  yield (api.test('Test qemu fail') + api.step_data(
      'Powerdown test',
      stdout=api.raw_io.output("""
              [No write since last change]
              Traceback (most recent call last):
              File \"/something/qemu/resources/qmp.py\", line 74, in <module>
                  main()
              File \"/something/qemu/resources/qmp.py\", line 58, in main
                  sock.connect(args.sock)
              FileNotFoundError: [Errno 2] No such file or directory
            """),
      retcode=1) + api.expect_exception('AssertionError') +
         api.post_process(DropExpectation))
