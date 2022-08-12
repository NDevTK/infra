# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import actions
from PB.recipes.infra.windows_image_builder import (online_windows_customization
                                                    as onlinewc)
from PB.recipes.infra.windows_image_builder import sources as sources
from PB.recipes.infra.windows_image_builder import dest as dest
from PB.recipes.infra.windows_image_builder import (windows_image_builder as
                                                    wib)
from PB.recipes.infra.windows_image_builder import vm as vm_pb
from PB.recipes.infra.windows_image_builder import actions as actions_pb
from PB.recipes.infra.windows_image_builder import windows_vm as windows_pb
from recipe_engine.post_process import DropExpectation, StatusFailure
from recipe_engine.post_process import StatusSuccess, StepCommandRE
from RECIPE_MODULES.infra.windows_scripts_executor import test_helper as t

DEPS = [
    'depot_tools/gsutil',
    'windows_scripts_executor',
    'recipe_engine/properties',
    'recipe_engine/platform',
    'recipe_engine/json',
    'recipe_engine/raw_io',
    'recipe_engine/path',
]

PYTHON_VERSION_COMPATIBILITY = 'PY3'

PROPERTIES = wib.Image


def RunSteps(api, config):
  api.windows_scripts_executor.init()
  custs = api.windows_scripts_executor.init_customizations(config)
  api.windows_scripts_executor.execute_customizations(custs)


def GenTests(api):
  image = 'Win10'
  cust = 'test'
  vm_name = 'Win10'

  INSTALL = t.VM_DRIVE(
      name='install',
      op=None,
      ip=sources.Src(
          gcs_src=sources.GCSSrc(
              bucket='ms-windows-images',
              source='Release/22621.1_MULTI_ARM64_EN-US.ISO')),
  )

  WIN_VM = t.VM_CONFIG(
      name=vm_name,
      drives=[INSTALL],
      device=[],
  )

  def IMAGE(arch):
    return t.WIN_IMAGE(
        image,
        arch,
        cust,
        vm_config=WIN_VM,
        action_list=[],
        win_config=windows_pb.WindowsVMConfig(boot_time=300, context={}))

  # Safely shutdown vm
  yield (api.test('Test shutdown_vm happy path') +
         api.platform('linux', 64, 'arm') +
         api.properties(IMAGE(wib.ARCH_AARCH64)) +
         # Mock successfully shutting down the vm
         t.SHUTDOWN_VM(api, image, cust, vm_name, 0) +
         # Mock vm status check. VM offline
         t.STATUS_VM(api, image, cust, vm_name) +
         # Recipe exits with success
         api.post_process(StatusSuccess) + api.post_process(DropExpectation))

  # The following events happen in this test
  # 1. Send shutdown vm to powershell [Fail]
  # 2. Send powerdown vm to QMP [Pass]
  # 3. Check vm status [Running] (VM OS ignored the powerdown signal)
  # 4. Send Quit vm to force quit [Success]
  # 5. Recipe fails
  # We force quit the VM. Cannot use the artifacts. Therefore the recipe fails
  yield (api.test('Test shutdown_vm fail (1)') +
         api.platform('linux', 64, 'arm') +
         api.properties(IMAGE(wib.ARCH_AARCH64)) +
         # Mock shutting down the vm fail
         t.SHUTDOWN_VM(api, image, cust, vm_name, 1) +
         # Mock powerdown vm successful
         t.POWERDOWN_VM(api, image, cust, vm_name) +
         # Mock vm status check. VM offline
         t.STATUS_VM(api, image, cust, vm_name, running=True) +
         # Force quit vm
         t.QUIT_VM(api, image, cust, vm_name) +
         # Recipe exits with Failure
         api.post_process(StatusFailure) + api.post_process(DropExpectation))

  # The following events happen in this test
  # 1. Send shutdown vm to powershell [Pass]
  # 2. Check vm status [Running] (Didn't shut down in time)
  # 3. Send Quit vm to force quit [Success]
  # 4. Recipe fails
  # We force quit the VM. Cannot use the artifacts. Therefore the recipe fails
  yield (api.test('Test shutdown_vm fail (2)') +
         api.platform('linux', 64, 'arm') +
         api.properties(IMAGE(wib.ARCH_AARCH64)) +
         # Mock shutting down the vm fail
         t.SHUTDOWN_VM(api, image, cust, vm_name, 1) +
         # Mock vm status check. VM offline
         t.STATUS_VM(api, image, cust, vm_name, running=True) +
         # Force quit vm
         t.QUIT_VM(api, image, cust, vm_name) +
         # Recipe exits with Failure
         api.post_process(StatusFailure) + api.post_process(DropExpectation))

  # The following events happen in this test
  # 1. Send shutdown vm to powershell [Fail]
  # 2. Send powerdown vm to QMP [Fail]
  # 3. Check vm status [Running] (Hope VM shuts down. But it didn't)
  # 4. Send Quit vm to force quit [Success]
  # 5. Recipe fails
  # We force quit the VM. Cannot use the artifacts. Therefore the recipe fails
  yield (api.test('Test shutdown_vm fail (3)') +
         api.platform('linux', 64, 'arm') +
         api.properties(IMAGE(wib.ARCH_AARCH64)) +
         # Mock shutting down the vm fail
         t.SHUTDOWN_VM(api, image, cust, vm_name, 1) +
         # Mock powerdown vm successful
         t.POWERDOWN_VM(api, image, cust, vm_name, False) +
         # Mock vm status check. VM offline
         t.STATUS_VM(api, image, cust, vm_name, running=True) +
         # Force quit vm
         t.QUIT_VM(api, image, cust, vm_name) +
         # Recipe exits with Failure
         api.post_process(StatusFailure) + api.post_process(DropExpectation))

  # The following events happen in this test
  # 1. Send shutdown vm to powershell [Fail]
  # 2. Send powerdown vm to QMP [Fail]
  # 3. Check vm status [Running] (Hope VM shuts down. But it didn't)
  # 4. Send Quit vm to force quit [Failure] (Cannot shut down vm. This is bad)
  # 5. Recipe fails
  # We failed. Cannot kill VM. Therefore the recipe fails
  yield (api.test('Test shutdown_vm fail (4)') +
         api.platform('linux', 64, 'arm') +
         api.properties(IMAGE(wib.ARCH_AARCH64)) +
         # Mock shutting down the vm fail
         t.SHUTDOWN_VM(api, image, cust, vm_name, 1) +
         # Mock powerdown vm successful
         t.POWERDOWN_VM(api, image, cust, vm_name, False) +
         # Mock vm status check. VM offline
         t.STATUS_VM(api, image, cust, vm_name, running=True) +
         # Force quit vm
         t.QUIT_VM(api, image, cust, vm_name, False) +
         # Recipe exits with Failure
         api.post_process(StatusFailure) + api.post_process(DropExpectation))
