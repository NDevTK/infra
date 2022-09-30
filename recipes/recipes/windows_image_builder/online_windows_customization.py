# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine import post_process

from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import actions
from PB.recipes.infra.windows_image_builder import sources
from PB.recipes.infra.windows_image_builder import windows_vm
from PB.recipe_engine.result import RawResult
from PB.go.chromium.org.luci.buildbucket.proto import common

from recipe_engine.post_process import DropExpectation, StatusSuccess
from RECIPE_MODULES.infra.windows_scripts_executor import test_helper as t

DEPS = [
    'depot_tools/gitiles',
    'recipe_engine/platform',
    'recipe_engine/properties',
    'recipe_engine/raw_io',
    'recipe_engine/json',
    'windows_adk',
    'windows_scripts_executor',
]

PYTHON_VERSION_COMPATIBILITY = 'PY3'

PROPERTIES = wib.Image


def RunSteps(api, image):
  """ This recipe executes offline_winpe_customization."""
  if api.platform.is_win:
    raise AssertionError('This recipe can only run on linux')

  # this recipe will only execute the online_windows_customization
  for cust in image.customizations:
    assert (cust.WhichOneof('customization') == 'online_windows_customization')

  # initialize the image to scripts executor
  api.windows_scripts_executor.init()

  custs = api.windows_scripts_executor.init_customizations(image)

  # pinning all the refs and generating unique keys
  custs = api.windows_scripts_executor.process_customizations(custs, {})

  # download all the required refs
  api.windows_scripts_executor.download_all_packages(custs)

  # execute the customizations given
  api.windows_scripts_executor.execute_customizations(custs)

  summary = 'Built:\n'
  for cust in custs:
    summary += '{}\n'.format(cust.get_key())
  # looks like everything executed properly, return result
  return RawResult(status=common.SUCCESS, summary_markdown=summary)


image = 'Win10'
cust = 'test'
vm_name = 'Win10VM'


def GenTests(api):

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

  ACTION_POWERSHELL_EXPR = actions.Action(
      powershell_expr=actions.PowershellExpr(
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
        action_list=[ACTION_POWERSHELL_EXPR],
        win_config=windows_vm.WindowsVMConfig(
            boot_time=300, context={'$deps_img': 'D:'}))

  yield (api.test('not_run_on_linux', api.platform('win', 64)) +
         api.expect_exception('AssertionError') +
         api.post_process(DropExpectation))

  yield (api.test('online windows customization happy path') +
         api.platform('linux', 64, 'arm') +
         api.properties(IMAGE(wib.ARCH_AARCH64)) +
         # Mock deps.img disk space check
         t.DISK_SPACE(api, image, cust, vm_name, 'deps.img') +
         # Mock disk mount for deps.img
         t.MOUNT_DISK(api, image, cust, vm_name, 'deps.img') +
         # Mock successful execution of powershell expression
         t.POWERSHELL_EXPR_VM(api, image, cust, 'Install HL3',
                              'HL3 installed successfully') +
         # Mock shutdown vm. successfully shut down vm
         t.SHUTDOWN_VM(api, image, cust, vm_name, 0) +
         # Mock stats vm check. VM offline
         t.STATUS_VM(api, image, cust, vm_name) +
         api.post_process(StatusSuccess) + api.post_process(DropExpectation))
