# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine.recipe_api import Property
from recipe_engine.post_process import StepCommandRE, DropExpectation
from os.path import exists

DEPS = ['depot_tools/git',
        'depot_tools/gsutil',
        'recipe_engine/archive',
        'recipe_engine/cipd',
        'recipe_engine/file',
        'recipe_engine/json',
        'recipe_engine/path',
        'recipe_engine/properties',
        'recipe_engine/step',
        'recipe_engine/time']

PROPERTIES = {
    'vm_count': Property(
        kind=int, help="Number of VMs to launch (1 or 2)", default=1),
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
  api.step('Check Cache', ['ls', api.path['cache']])

  api.step('Check Cache Again', ['ls', api.path['cache']])
  cache_dir = api.path['cache']
  binary_loc = cache_dir.join("Purple_Panda-Swift.app","Contents","MacOS","Purple_Panda-Swift")

  bundle_zip = str(api.path['cache'].join("VM.bundle.zip"))
  bundle = str(api.path['cache'].join("VM.bundle"))

  # Download the vm bundle from google cloud storage if not already in cache
  if not exists(bundle_zip):
    api.gsutil.download("mac-vm-images", "VM.bundle.zip", api.path['cache'], args=None)

  # Only unzip if needed
  if not exists(bundle):
    api.archive.extract("Unzip VM Bundle", bundle_zip, api.path['cache'])

  api.step('Check Cache A Third Time', ['ls', api.path['cache']])

  if vm_count == 1:
    api.step("VM Launching Script 1", ["vpython3", "-u", api.resource("mac_vm_script.py"), "--binary_path", binary_loc, "--cpu_count", 4, "--bundle", bundle])

  # api.step("Launch VM With Open Command", ["open", binary_loc, "--args", "4", "4"])

  api.time.sleep(60*60*4)
#
#  else:
#    api.step("VM Launching Script 2", ["vpython3", "-u", api.resource("mac_vm_script.py"), "--binary_path", binary_loc, "--cpu_count", 4, "--bundle", vm1_bundle, vm2_bundle, "-d"])

def GenTests(api):
 yield api.test('basic') + api.step_data('VM Launching Script', retcode=0)
 yield api.test(
   'One VM',
   api.properties(vm_count = 1, vm1_bundle="dummy_loc", vm2_bundle="dummy_loc"),
   api.post_process(StepCommandRE, "VM Launching Script 1",
        ["vpython3", "-u", "RECIPE[infra::vm_launcher/vm_launcher].resources/mac_vm_script.py", "--binary_path", "[CACHE]/Purple_Panda-Swift.app/Contents/MacOS/Purple_Panda-Swift", "--cpu_count", "4", "--bundle", "dummy_loc"]),
      api.post_process(DropExpectation)
  ) + api.step_data('VM Launching Script', retcode=0)

 yield api.test(
   'Two VMs',
   api.properties(vm_count = 2, vm1_bundle="dummy_loc", vm2_bundle="dummy_loc"),
   api.post_process(StepCommandRE, "VM Launching Script 2",
        ["vpython3", "-u", "RECIPE[infra::vm_launcher/vm_launcher].resources/mac_vm_script.py", "--binary_path", "[CACHE]/Purple_Panda-Swift.app/Contents/MacOS/Purple_Panda-Swift", "--cpu_count", "4", "--bundle", "dummy_loc", "dummy_loc", "-d"]),
      api.post_process(DropExpectation)
  ) + api.step_data('VM Launching Script', retcode=0)
