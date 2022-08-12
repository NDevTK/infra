# Copyright 2019 The Chromium Authors. All rights reserved.
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
  custs = api.windows_scripts_executor.process_customizations(custs)
  api.windows_scripts_executor.execute_customizations(custs)


def GenTests(api):
  image = 'Win10'
  cust = 'test'
  vm_name = 'Win10'

  INSTALL = t.VM_DRIVE(
      name='install.iso',
      op=None,
      ip=sources.Src(
          gcs_src=sources.GCSSrc(
              bucket='ms-windows-images',
              source='Release/22621.1_MULTI_ARM64_EN-US.ISO')),
  )

  WIN_VM = t.VM_CONFIG(
      name=vm_name,
      drives=[INSTALL],
      device=['ide-cd,drive=install.iso'],
  )

  ACTION_ADD_FILE = actions_pb.Action(
      powershell_expr=actions_pb.PowershellExpr(
          name='Install HL3',
          expr='python $install_hl3_py',
          srcs={
              'install_hl3_py':
                  sources.Src(
                      cipd_src=sources.CIPDSrc(
                          package='infra/software/hl3',
                          refs='stable',
                          platform='windows-arm64'))
          },
      ),)

  def IMAGE(arch):
    return t.WIN_IMAGE(
        image,
        arch,
        cust,
        vm_config=WIN_VM,
        action_list=[ACTION_ADD_FILE],
        win_config=windows_pb.WindowsVMConfig(
            boot_time=300, context={
                '$system_img': 'C:',
                '$deps_img': 'D:'
            }))

  yield (api.test('powershell_expr test happy path') +
         api.platform('linux', 64, 'arm') +
         api.properties(IMAGE(wib.ARCH_AARCH64)) +
         # Mock deps.img disk space check
         t.DISK_SPACE(api, image, cust, vm_name, 'deps.img') +
         # Mock disk mount for deps.img
         t.MOUNT_DISK(api, image, cust, vm_name, 'deps.img') +
         # Mock successful execution of powershell expression
         t.POWERSHELL_EXPR_VM(api, image, cust, 'Install HL3',
                              'HL3 installed successfully') +
         api.post_process(StatusSuccess) + api.post_process(DropExpectation))

  yield (api.test('powershell_expr test fail') +
         api.platform('linux', 64, 'intel') +
         api.properties(IMAGE(wib.ARCH_AARCH64)) +
         # Mock deps.img disk space check
         t.DISK_SPACE(api, image, cust, vm_name, 'deps.img') +
         # Mock disk mount for deps.img
         t.MOUNT_DISK(api, image, cust, vm_name, 'deps.img') +
         # Mock failed execution of powershell expression
         t.POWERSHELL_EXPR_VM(
             api,
             image,
             cust,
             'Install HL3',
             '',
             'Failed to install, not arm device',
             retcode=12,
             success=False) + api.post_process(StatusFailure) +
         api.post_process(DropExpectation))
