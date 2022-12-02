# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.post_process import StepCommandRE, DropExpectation

PYTHON_VERSION_COMPATIBILITY = 'PY3'

DEPS = ['recipe_engine/step',
        'recipe_engine/cipd',
        'recipe_engine/path']

def RunSteps(api):
  # api.step('Print Hello World', ['echo', 'hello', 'world'])
  ef = api.cipd.EnsureFile()
  ef.add_package(name='experimental/jairogarciga_at_google.com/purple_panda',
                 version='latest')                          # Add the purple panda binary
  api.cipd.ensure(root=api.path['cache'], ensure_file=ef)   # Ensures that the binary is loaded into the cache path
  api.step('Check what we have', ['ls', api.path['cache']]) # Show the files in the cache

  vm_launcher = "~/cr/infra_internal/infra_internal/tools/purple_panda/Python/mac_vm_script.py"

  cache_dir = api.path['cache']
  binary_loc = str(cache_dir) + "/purple_panda"

  api.step('Print Hello World', ['echo', binary_loc])
  api.step('Access the VM launching script', ["python3", vm_launcher," --binary_path", "binary_loc",
                                                "--cpu_count", 4, "--bundle", "vm_bundle_loc"])

def GenTests(api):
 yield api.test('basic')
