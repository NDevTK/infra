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
  custs = api.windows_scripts_executor.process_customizations(custs, {})
  custs = api.windows_scripts_executor.filter_executable_customizations(custs)
  api.windows_scripts_executor.execute_customizations(custs)


def GenTests(api):
  image = 'Win10'
  cust = 'test'
  vm_name = 'Win10'
  boot_time = 360

  SYSTEM = t.VM_DRIVE(
      name='system',
      ip=sources.Src(
          gcs_src=sources.GCSSrc(
              bucket='chrome-windows-images',
              source='windows_images/system.zip')),
      op=[
          dest.Dest(
              gcs_src=sources.GCSSrc(
                  bucket='chrome-windows-images', source='WIN-OUT/system'))
      ],
      size=10240,
      media='disk',
      filesystem='fat',
      interface='none')

  INSTALL = t.VM_DRIVE(
      name='install',
      op=None,
      readonly=True,
      ip=sources.Src(
          gcs_src=sources.GCSSrc(
              bucket='ms-windows-images',
              source='Release/22621.1_MULTI_ARM64_EN-US.ISO')),
  )

  WIN_VM = t.VM_CONFIG(
      name=vm_name,
      drives=[SYSTEM, INSTALL],
      device=['-device ide-hd,drive=system'],
  )

  ACTION_ADD_FILE = actions_pb.Action(
      add_file=actions_pb.AddFile(
          name='Bootstrap example.py',
          src=sources.Src(
              cipd_src=sources.CIPDSrc(
                  package='infra/tools/example',
                  refs='stable',
                  platform='windows-amd64')),
          dst='$system',
      ))

  def IMAGE(arch, mode=wib.CustomizationMode.CUST_NORMAL):
    return t.WIN_IMAGE(
        image,
        arch,
        cust,
        vm_config=WIN_VM,
        action_list=[ACTION_ADD_FILE],
        win_config=windows_pb.WindowsVMConfig(
            boot_time=boot_time,
            context={
                '$system': 'C:',
                '$DEPS': 'D:',
            },
            shutdown_time=300),
        mode=mode,
    )

  key_win = '58d14c6fc3a92d22be294beda85d0a471c70af02dad2cfddfa80626ac1604d12'
  system = 'boot(windows_cust)-drive(system)-output.zip'

  yield (api.test('execute_customization_happy_path[AARCH64_KVM]') +
         api.platform('linux', 64, 'arm') +
         api.properties(IMAGE(wib.ARCH_AARCH64)) +
         t.DISK_SPACE(api, image, cust, vm_name, 'DEPS') +
         t.MOUNT_DISK(api, image, cust, vm_name, 'DEPS') +
         t.STARTUP_VM(api, image, cust, vm_name, True) +
         t.ADD_FILE_VM(api, image, cust, 'Bootstrap example.py', 1) +
         t.SHUTDOWN_VM(api, image, cust, vm_name, 1) +
         t.STATUS_VM(api, image, cust, vm_name) + t.MOCK_CUST_OUTPUT(
             api, 'gs://chrome-gce-images/WIB-ONLINE-CACHE/{}-{}'.format(
                 key_win, system), False) + api.post_process(StatusSuccess) +
         api.post_process(DropExpectation))

  yield (api.test('execute_customization_happy_path[AMD64_KVM]') +
         api.platform('linux', 64, 'intel') +
         api.properties(IMAGE(wib.ARCH_AMD64)) +
         t.DISK_SPACE(api, image, cust, vm_name, 'DEPS') +
         t.MOUNT_DISK(api, image, cust, vm_name, 'DEPS') +
         t.STARTUP_VM(api, image, cust, vm_name, True) +
         t.ADD_FILE_VM(api, image, cust, 'Bootstrap example.py', 1) +
         t.SHUTDOWN_VM(api, image, cust, vm_name, 0) +
         t.STATUS_VM(api, image, cust, vm_name) + t.MOCK_CUST_OUTPUT(
             api, 'gs://chrome-gce-images/WIB-ONLINE-CACHE/{}-{}'.format(
                 key_win, system), False) + api.post_process(StatusSuccess) +
         api.post_process(DropExpectation))

  yield (api.test('execute_customization_happy_path[X86_KVM]') +
         api.platform('linux', 32, 'intel') +
         api.properties(IMAGE(wib.ARCH_X86)) +
         t.DISK_SPACE(api, image, cust, vm_name, 'DEPS') +
         t.MOUNT_DISK(api, image, cust, vm_name, 'DEPS') +
         t.STARTUP_VM(api, image, cust, vm_name, True) +
         t.ADD_FILE_VM(api, image, cust, 'Bootstrap example.py', 1) +
         t.SHUTDOWN_VM(api, image, cust, vm_name, 0) +
         t.STATUS_VM(api, image, cust, vm_name) + t.MOCK_CUST_OUTPUT(
             api, 'gs://chrome-gce-images/WIB-ONLINE-CACHE/{}-{}'.format(
                 key_win, system), False) + api.post_process(StatusSuccess) +
         api.post_process(DropExpectation))

  yield (api.test('execute_customization_fail_add_file') +
         api.platform('linux', 64, 'intel') +
         api.properties(IMAGE(wib.ARCH_AMD64)) +
         t.DISK_SPACE(api, image, cust, vm_name, 'DEPS') +
         t.MOUNT_DISK(api, image, cust, vm_name, 'DEPS') + t.ADD_FILE_VM(
             api, image, cust, 'Bootstrap example.py', 8, success=False) +
         t.STARTUP_VM(api, image, cust, vm_name, True) +
         t.SHUTDOWN_VM(api, image, cust, vm_name, 0) +
         t.STATUS_VM(api, image, cust, vm_name) + t.MOCK_CUST_OUTPUT(
             api, 'gs://chrome-gce-images/WIB-ONLINE-CACHE/{}-{}'.format(
                 key_win, system), False) + api.expect_status('FAILURE') +
         api.post_process(DropExpectation))

  yield (api.test('execute_customization_fail_add_file_debug') +
         api.platform('linux', 64, 'intel') +
         # enable debug mode
         api.properties(
             IMAGE(wib.ARCH_AMD64, mode=wib.CustomizationMode.CUST_DEBUG)) +
         t.DISK_SPACE(api, image, cust, vm_name, 'DEPS') +
         t.MOUNT_DISK(api, image, cust, vm_name, 'DEPS') + t.ADD_FILE_VM(
             api, image, cust, 'Bootstrap example.py', 8, success=False) +
         t.STARTUP_VM(api, image, cust, vm_name, True) +
         t.SHUTDOWN_VM(api, image, cust, vm_name, 0) +
         t.CHECK_DEBUG_SLEEP(api, image, cust, time=boot_time) +
         t.STATUS_VM(api, image, cust, vm_name) + t.MOCK_CUST_OUTPUT(
             api, 'gs://chrome-gce-images/WIB-ONLINE-CACHE/{}-{}'.format(
                 key_win, system), False) + api.expect_status('FAILURE') +
         api.post_process(DropExpectation))

  yield (api.test('execute_customization_fail_safe_shutdown') +
         api.platform('linux', 64, 'intel') +
         api.properties(IMAGE(wib.ARCH_AMD64)) +
         t.DISK_SPACE(api, image, cust, vm_name, 'DEPS') +
         t.MOUNT_DISK(api, image, cust, vm_name, 'DEPS') + t.ADD_FILE_VM(
             api, image, cust, 'Bootstrap example.py', 8, success=False) +
         t.STARTUP_VM(api, image, cust, vm_name, True) +
         t.SHUTDOWN_VM(api, image, cust, vm_name, 0) +
         t.STATUS_VM(api, image, cust, vm_name, running=True) +
         t.QUIT_VM(api, image, cust, vm_name, success=True) +
         t.MOCK_CUST_OUTPUT(
             api, 'gs://chrome-gce-images/WIB-ONLINE-CACHE/{}-{}'.format(
                 key_win, system), False) + api.expect_status('FAILURE') +
         api.post_process(DropExpectation))
