# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.recipe_api import Property
from recipe_engine.post_process import StepCommandRE, DropExpectation

DEPS = ['depot_tools/git',
        'recipe_engine/cipd',
        'recipe_engine/json',
        'recipe_engine/path',
        'recipe_engine/properties',
        'recipe_engine/step']

PROPERTIES = {
    'vm_count': Property(
        kind=str, help="Number of VMs to launch (1 or 2)", default=1),
    'vm1_bundle': Property(
        kind=str, help="The location of the first binary", default="dummy_loc"),
    'vm2_bundle': Property(
        kind=str, help="The location of the second binary", default="dummy_loc2"),
}

def RunSteps(api, vm_count, vm1_bundle, vm2_bundle):

  ef = api.cipd.EnsureFile()
  ef.add_package(name='experimental/jairogarciga_at_google.com/purple_panda',
                 version='latest')
                 # Add the purple panda binary

  # Ensures that the binary is loaded into the cache path
  api.cipd.ensure(root=api.path['cache'], ensure_file=ef)
  api.step('Check what we have', ['ls', api.path['cache']])

  api.git.checkout(url="https://chrome-internal.googlesource.com/infra/infra_internal",
      dir_path=api.path['cache'],
      file_name="infra/infra_internal/+/refs/heads/main/infra_internal/tools/purple_panda/Python/mac_vm_script.py")



  api.step('Check Again', ['ls', api.path['cache']])
  cache_dir = api.path['cache']
  binary_loc = cache_dir.join("Purple_Panda-Swift")

  api.step('Print Hello World', ['echo', binary_loc])

  vm_launcher = api.path['cache'].join("mac_vm_script.py")

  if vm_count == "1":
    api.step('Single VM', ["python3", vm_launcher," --binary_path", binary_loc, "--cpu_count", 4, "--bundle", vm1_bundle])
  else:
    api.step('Dual VMs', ["python3", vm_launcher," --binary_path", "binary_loc", "--cpu_count", 4, "--bundle", vm1_bundle, vm2_bundle, "-d"])

def GenTests(api):
 yield api.test('basic')
 yield api.test(
   'One VM',
   api.properties(vm_count = "1", vm1_bundle="dummy_loc", vm2_bundle="dummy_loc"),
   api.post_process(StepCommandRE, "Single VM",
        ["python3", "[CACHE]/mac_vm_script.py"," --binary_path", "[CACHE]/Purple_Panda-Swift", "--cpu_count", "4", "--bundle", "dummy_loc"]),
      api.post_process(DropExpectation)
  )
 yield api.test(
   'Two VMs',
   api.properties(vm_count = "2", vm1_bundle="dummy_loc", vm2_bundle="dummy_loc"),
   api.post_process(StepCommandRE, "Dual VMs",
        ["python3", "[CACHE]/mac_vm_script.py"," --binary_path", "[CACHE]/Purple_Panda-Swift", "--cpu_count", "4", "--bundle", "dummy_loc", "dummy_loc", "-d"]),
      api.post_process(DropExpectation)
  )
