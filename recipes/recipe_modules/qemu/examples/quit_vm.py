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
  api.qemu.quit_vm(name='test')


def GenTests(api):

  # Test successful execution of quit_vm
  yield (api.test('Test quit vm') + api.post_process(StatusSuccess) +
         api.step_data('Quit test', api.json.output({
             'return': {},
         })) + api.post_process(DropExpectation))

  # Test failure to quit_vm, No exception thrown
  yield (api.test('Test quit vm fail') + api.post_process(StatusSuccess) +
         api.step_data(
             'Quit test',
             api.raw_io.output("""
              [No write since last change]
              Traceback (most recent call last):
              File \"/something/qemu/resources/qmp.py\", line 74, in <module>
                  main()
              File \"/something/qemu/resources/qmp.py\", line 58, in main
                  sock.connect(args.sock)
              FileNotFoundError: [Errno 2] No such file or directory
            """),
             retcode=1) + api.post_process(DropExpectation))
