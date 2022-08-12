# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import (offline_winpe_customization
                                                    as winpe)
from PB.recipes.infra.windows_image_builder import actions
from PB.recipes.infra.windows_image_builder import sources
from PB.recipes.infra.windows_image_builder import dest

from recipe_engine.post_process import DropExpectation, StatusFailure
from recipe_engine.post_process import StatusSuccess, StepCommandRE
from RECIPE_MODULES.infra.windows_scripts_executor import test_helper as t

DEPS = [
    'depot_tools/gitiles',
    'recipe_engine/properties',
    'recipe_engine/platform',
    'recipe_engine/json',
    'recipe_engine/path',
    'recipe_engine/raw_io',
    'windows_scripts_executor',
]

PYTHON_VERSION_COMPATIBILITY = 'PY3'

PROPERTIES = wib.Image


def RunSteps(api, config):
  api.windows_scripts_executor.init()
  custs = api.windows_scripts_executor.init_customizations(config)
  api.windows_scripts_executor.process_customizations(custs)
  api.windows_scripts_executor.download_all_packages(custs)


def GenTests(api):
  ACTION_ADD_BOOTSTRAP = actions.Action(
      add_file=actions.AddFile(
          name='add_bootstrap_file',
          src=sources.Src(
              git_src=sources.GITSrc(
                  repo='chromium.dev',
                  ref='HEAD',
                  src='windows/artifacts/bootstrap.ps1'),),
          dst='C:\\Windows\\System32',
      ))

  SYSTEM = t.VM_DRIVE(
      name='system',
      ip=None,
      op=dest.Dest(
          gcs_src=sources.GCSSrc(
              bucket='chrome-windows-images', source='WIN-OUT/system.img')),
  )

  INSTALL = t.VM_DRIVE(
      name='install',
      op=None,
      ip=sources.Src(
          gcs_src=sources.GCSSrc(
              bucket='ms-windows-images',
              source='Release/22621.1_MULTI_ARM64_EN-US.ISO')),
  )

  AARCH64_VM = t.VM_CONFIG(
      name='WinArm', version='latest', drives=[SYSTEM, INSTALL])

  yield (api.test('pin_download_all_deps') + api.properties(
      t.WIN_IMAGE(
          'WinArm',
          wib.ARCH_AARCH64,
          'test',
          vm_config=AARCH64_VM,
          action_list=[ACTION_ADD_BOOTSTRAP])) + t.GIT_PIN_FILE(
              api, 'test', 'HEAD', 'windows/artifacts/bootstrap.ps1', 'HEAD') +
         api.post_process(DropExpectation))
