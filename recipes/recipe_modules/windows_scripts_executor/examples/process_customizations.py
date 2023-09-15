# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import windows_iso as winiso
from PB.recipes.infra.windows_image_builder import sources

from recipe_engine.post_process import DropExpectation, StatusFailure
from recipe_engine.post_process import StatusSuccess
from RECIPE_MODULES.infra.windows_scripts_executor import test_helper as t

DEPS = [
    'recipe_engine/properties', 'recipe_engine/platform',
    'windows_scripts_executor'
]

PYTHON_VERSION_COMPATIBILITY = 'PY3'

PROPERTIES = wib.Image


def RunSteps(api, config):
  api.windows_scripts_executor.init()
  custs = api.windows_scripts_executor.init_customizations(config)
  # Get all the inputs required. This will be used to determine if we have
  # to cache any images in online customization
  inputs = []
  for cust in custs:
    for ip in cust.inputs:
      if ip.WhichOneof('src') == 'local_src':
        inputs.append(ip.local_src)

  custs = api.windows_scripts_executor.process_customizations(custs, {}, inputs)


def GenTests(api):
  # Test context resolution by referencing a local customization. This provides
  # two customizations with one referencing the other.
  image = 'process_customizations_test'
  arch = wib.ARCH_AMD64
  cust1 = 'cust_1'
  cust2 = 'cust_2'
  sub_c1 = 'subcustomization_1'
  sub_c2 = 'subcustomiaztion_2'
  REF_IMAGE = t.WPE_IMAGE(image, arch, cust1, sub_c1, action_list=[])
  IMAGE_REFERENCING_PREV_IMAGE = t.WPE_IMAGE(
      image,
      arch,
      cust2,
      sub_c2,
      action_list=[],
      image_src=sources.Src(
          local_src='image(process_customizations_test)-cust(cust_1)-output'))
  # Add the second customization to the first image. Now there are two
  # customizations in the REF_IMAGE with second referencing the first one.
  REF_IMAGE.customizations.extend(IMAGE_REFERENCING_PREV_IMAGE.customizations)

  yield (api.test('Test process_customizations Happy Path',
                  api.platform('win', 64)) +
         # input image with install file action without any args
         api.properties(REF_IMAGE) + api.post_process(StatusSuccess) +
         api.post_process(DropExpectation))

  # Test context resolution cyclical dependency. We create two customizations
  # with the first one referencing the second and the second one referencing the
  # first
  REF_IMAGE = t.WPE_IMAGE(
      image,
      arch,
      cust1,
      sub_c1,
      action_list=[],
      image_src=sources.Src(local_src='process_customizations_test-cust_2'))
  # Add the second customization to the first image. Now there are two
  # customizations in the REF_IMAGE with second referencing the first one and
  # the first referencing the second one. This should fail
  REF_IMAGE.customizations.extend(IMAGE_REFERENCING_PREV_IMAGE.customizations)

  yield (api.test('Test process_customizations Cyclical dependency',
                  api.platform('win', 64)) +
         # input image with install file action without any args
         api.properties(REF_IMAGE) + api.expect_status('FAILURE') +
         api.post_process(DropExpectation))


  # Test dependency injection for online windows customizations. Create a
  # Online customization without any outputs but reference the output in
  # another customization. This should trigger the dependency injection for the
  # cust.

  # Create a system drive without any explicit dests
  SYSTEM = t.VM_DRIVE(
      name='system',
      ip=None,
      op=None,
  )

  # Disk to install the OS on system drive
  INSTALL = t.VM_DRIVE(
      name='install',
      op=None,
      ip=sources.Src(
          gcs_src=sources.GCSSrc(
              bucket='ms-windows-images',
              source='Release/22621.1_MULTI_ARM64_EN-US.ISO')),
  )

  # Add a VM config for this online cust
  AMD64_VM = t.VM_CONFIG(
      name='WinVM', version='latest', drives=[SYSTEM, INSTALL])

  # INJ_TEST is the image that will test dependency injection
  ON_CUST = t.WIN_IMAGE(
      'inj_test',
      wib.ARCH_AMD64,
      'online_cust',
      vm_config=AMD64_VM,
      action_list=[])

  SYSTEM = 'image(inj_test)-cust(online_cust)'+\
      '-boot(windows_cust)-drive(system)-output'
  ISO_CUST = t.WIN_ISO(
      image='inj_test',
      arch=wib.ARCH_AMD64,
      name='iso_cust',
      copy_files=[
          winiso.CopyArtifact(
              artifact=sources.Src(local_src=SYSTEM,),
              mount=True,
              source='install.wim',
              dest='sources/install.wim')
      ])

  ON_CUST.customizations.extend(ISO_CUST.customizations)

  yield (api.test('Deps injection', api.platform('linux', 64)) +
         # input image with install file action without any args
         api.properties(ON_CUST) + api.post_process(StatusSuccess) +
         api.post_process(DropExpectation))
