# Copyright 2022 The Chromium Authors
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
  api.windows_scripts_executor.download_all_packages(custs)
  api.path.mock_add_paths('[CLEANUP]/gen_iso/workdir/staging/gen_iso.iso',
                          'FILE')
  api.path.mock_add_paths(
      '[CACHE]/Pkgs/CIPDPkgs/resolved-instance_id-of-latest----------'
      '/infra/chrome/windows/wallpapers/windows-amd64', 'DIRECTORY')
  api.path.mock_add_paths(
      '[CACHE]/Pkgs/CIPDPkgs/resolved-instance_id-of-latest----------'
      '/infra/chrome/nopompt_boot/windows-amd64', 'DIRECTORY')
  api.windows_scripts_executor.execute_customizations(custs)


def GenTests(api):
  image = 'test_iso_generation'
  cust = 'gen_iso'
  arch = wib.ARCH_AMD64
  key = "6d1f9fdab5c27a6b86389c89ed3dce53067eaadcfd7b5e8fc18028a59c00b0b1"
  uploads = [
      dest.Dest(
          gcs_src=sources.GCSSrc(
              bucket='chrome-windows-image', source='Win10/win10_release.iso'),
          tags={'VERSION': '10'},
      )
  ]

  yield (api.test('Happy path with custom bootloader') +
         api.platform('linux', 64, 'intel') + api.properties(
             t.WIN_ISO(image=image, arch=arch, name=cust, uploads=uploads)) +
         t.MOCK_CUST_OUTPUT(
             api, "gs://chrome-gce-images/WIB-ISO/{}.iso".format(key), False) +
         t.MOUNT_DISK_ISO(api, image, cust,
                          'gs://chrome-gce-images/WIN-ISO/win10_vanilla.iso') +
         t.MOUNT_DISK_ISO(api, image, cust,
                          'gs://chrome-gce-images/WIB-ONLINE-CACHE/st.zip') +
         t.MOUNT_DISK_ISO(api, image, cust,
                          'gs://chrome-gce-images/WIN-CACHE/win10_chrome.iso') +
         api.post_process(StatusSuccess) + api.post_process(DropExpectation))

  yield (api.test('Happy path with default bootloader') +
         api.platform('linux', 64, 'intel') + api.properties(
             t.WIN_ISO(
                 image=image,
                 arch=arch,
                 name=cust,
                 uploads=uploads,
                 boot_image=None)) +
         t.MOCK_CUST_OUTPUT(
             api, "gs://chrome-gce-images/WIB-ISO/{}.iso".format(key), False) +
         t.MOUNT_DISK_ISO(api, image, cust,
                          'gs://chrome-gce-images/WIN-ISO/win10_vanilla.iso') +
         t.MOUNT_DISK_ISO(api, image, cust,
                          'gs://chrome-gce-images/WIB-ONLINE-CACHE/st.zip') +
         t.MOUNT_DISK_ISO(api, image, cust,
                          'gs://chrome-gce-images/WIN-CACHE/win10_chrome.iso') +
         api.post_process(StatusSuccess) + api.post_process(DropExpectation))
