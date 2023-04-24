# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine import post_process

from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import actions
from PB.recipes.infra.windows_image_builder import sources
from PB.recipes.infra.windows_image_builder import dest
from PB.recipes.infra.windows_image_builder import windows_vm
from PB.recipes.infra.windows_image_builder import windows_iso as winiso
from PB.recipe_engine.result import RawResult
from PB.go.chromium.org.luci.buildbucket.proto import common

from recipe_engine.post_process import DropExpectation, StatusSuccess
from recipe_engine.post_process import StatusFailure
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
    assert (
        cust.WhichOneof('customization') == 'online_windows_customization') or (
            cust.WhichOneof('customization') == 'windows_iso_customization')

  # initialize the image to scripts executor
  api.windows_scripts_executor.init()

  custs = api.windows_scripts_executor.init_customizations(image)

  # Get all the inputs required. This will be used to determine if we have
  # to cache any images in online customization
  inputs = []
  for cust in custs:
    for ip in cust.inputs:
      if ip.WhichOneof('src') == 'local_src':
        inputs.append(ip.local_src)

  # pinning all the refs and generating unique keys
  custs = api.windows_scripts_executor.process_customizations(custs, {}, inputs)

  # Execute customizations one by one
  built_custs = []
  failed_custs_errs = []
  while len(built_custs) + len(failed_custs_errs) != len(custs):
    to_build = []
    # List of failed customizations without errors
    failed_custs = [c for c, _ in failed_custs_errs]
    for cust in custs:
      if cust not in built_custs and cust not in failed_custs:
        if cust.executable():
          to_build.append(cust)
    if to_build:
      # We can't parallelize this, because we need to ensure that these
      # get executed in order. Otherwise we might fail when we try to
      # fetch something that doesn't exist yet.
      for to_exec in to_build:
        try:
          # download all the required refs
          api.windows_scripts_executor.download_all_packages([to_exec])
          # execute the customizations given
          api.windows_scripts_executor.execute_customizations([to_exec])
          built_custs.extend([to_exec])
        except Exception as e:
          # Collect failed custs and attempt to execute others
          failed_custs_errs.append((to_exec, e))
    else:
      # We are done building. We can't build anything else
      break  # pragma: nocover
  # Customizations that we couldn't execute
  couldnot_exec = []
  f_custs = [cust for cust, _ in failed_custs_errs]
  for cust in custs:
    if (cust not in f_custs) and (cust not in built_custs):
      couldnot_exec.append(cust)
  # Report back
  summary = ''
  if built_custs:
    summary += 'Built: <br>'
    for cust in built_custs:
      summary += '{}--{} <br>'.format(cust.id, cust.get_key())
  if couldnot_exec:
    summary += 'Did not build: <br>'
    for cust in couldnot_exec:
      summary += '{}--{} <br>'.format(cust.id, cust.get_key())
  if failed_custs_errs:
    summary += 'Failed: <br>'
    for cust, err in failed_custs_errs:
      summary += '{}--{}: {} <br>'.format(cust.id, cust.get_key(), err)
  status = common.SUCCESS
  if len(failed_custs_errs) + len(couldnot_exec) > 0:
    status = common.FAILURE
  # looks like everything executed properly, return result
  return RawResult(status=status, summary_markdown=summary)


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

  SYSTEM = t.VM_DRIVE(
      name='system',
      op=[
          dest.Dest(
              gcs_src=sources.GCSSrc(
                  bucket='chrome-gce-images',
                  source='tests/sys.img',
              ))
      ],
      ip=None)

  WIN_VM = t.VM_CONFIG(
      name=vm_name,
      drives=[INSTALL, SYSTEM],
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
            boot_time=300, context={'$DEPS': 'D:'}))

  cust_key = '58d14c6fc3a92d22be294beda85d0a471c70af02dad2cfddfa80626ac1604d12'
  cache = '{}-boot(windows_cust)-drive(system)-output.zip'.format(cust_key)

  yield (api.test('not_run_on_linux', api.platform('win', 64)) +
         api.expect_exception('AssertionError') +
         api.post_process(DropExpectation))

  yield (api.test('online windows customization happy path') +
         api.platform('linux', 64, 'arm') +
         api.properties(IMAGE(wib.ARCH_AARCH64)) +
         # Mock DEPS disk space check
         t.DISK_SPACE(api, image, cust, vm_name, 'DEPS') +
         # Mock disk mount for DEPS
         t.MOUNT_DISK(api, image, cust, vm_name, 'DEPS') +
         # Mock DEPS disk space check
         t.DISK_SPACE(api, image, cust, vm_name, 'system', size=68719486736) +
         # Mock the vm start/boot
         t.STARTUP_VM(api, image, cust, vm_name, True) +
         # Mock successful execution of powershell expression
         t.POWERSHELL_EXPR_VM(api, image, cust, 'Install HL3',
                              'HL3 installed successfully') +
         t.MOCK_CUST_OUTPUT(
             api, 'gs://chrome-gce-images/WIB-ONLINE-CACHE/{}'.format(cache),
             False) +
         # Mock shutdown vm. successfully shut down vm
         t.SHUTDOWN_VM(api, image, cust, vm_name, 0) +
         # Mock stats vm check. VM offline
         t.STATUS_VM(api, image, cust, vm_name) +
         api.post_process(StatusSuccess) + api.post_process(DropExpectation))

  IMG = IMAGE(wib.ARCH_AARCH64)
  # system drive from online cust
  SYSTEM = 'image(Win10)-cust(test)-boot(windows_cust)-drive(system)-output'
  # Add a related Winiso customization
  IMG.customizations.extend(
      t.WIN_ISO(
          image=image,
          arch=wib.ARCH_AARCH64,
          name='cdrom',
          uploads=[],
          copy_files=[
              winiso.CopyArtifact(
                  # copy wim from system drive
                  artifact=sources.Src(local_src=SYSTEM),
                  mount=True,
                  source='install.wim',
                  dest='sources/install.wim')
          ],
      ).customizations)

  yield (
      api.test('online windows customization fail') +
      api.platform('linux', 64, 'arm') + api.properties(IMG) +
      # Mock DEPS disk space check
      t.DISK_SPACE(api, image, cust, vm_name, 'DEPS') +
      # Mock disk mount for DEPS
      t.MOUNT_DISK(api, image, cust, vm_name, 'DEPS') +
      # Mock DEPS disk space check
      t.DISK_SPACE(api, image, cust, vm_name, 'system', size=68719486736) +
      # Mock the vm start/boot
      t.STARTUP_VM(api, image, cust, vm_name, True) +
      # Mock successful execution of powershell expression
      t.POWERSHELL_EXPR_VM(
          api, image, cust, 'Install HL3', 'HL3 failed install',
          success=False) + t.MOCK_CUST_OUTPUT(
              api, 'gs://chrome-gce-images/WIB-ONLINE-CACHE/{}'.format(cache),
              False) +
      # Mock shutdown vm. successfully shut down vm
      t.SHUTDOWN_VM(api, image, cust, vm_name, 0) +
      # Mock stats vm check. VM offline
      t.STATUS_VM(api, image, cust, vm_name) + api.post_process(StatusFailure) +
      api.post_process(DropExpectation))
