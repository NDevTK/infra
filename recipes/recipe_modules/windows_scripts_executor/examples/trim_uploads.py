# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import sources
from PB.recipes.infra.windows_image_builder import dest

from recipe_engine.post_process import DropExpectation
from recipe_engine.post_process import StatusSuccess
from RECIPE_MODULES.infra.windows_scripts_executor import test_helper as t

DEPS = [
    'windows_scripts_executor', 'recipe_engine/path',
    'recipe_engine/properties', 'recipe_engine/platform', 'recipe_engine/json',
    'recipe_engine/raw_io'
]

PYTHON_VERSION_COMPATIBILITY = 'PY3'

PROPERTIES = wib.Image


def RunSteps(api, config):
  api.windows_scripts_executor.init()
  custs = api.windows_scripts_executor.init_customizations(config)
  api.windows_scripts_executor.trim_uploads(custs)
  for cust in custs:
    assert (not cust.outputs)


def GenTests(api):
  WINPE_IMAGE = t.WPE_IMAGE(
      "test",
      wib.ARCH_X86,
      "wpe",
      "nop", [],
      up_dests=[
          dest.Dest(
              cipd_src=sources.CIPDSrc(
                  package='experimental/mock/wib/test-1',
                  refs='latest',
                  platform='windows-amd64',
              ),
              tags={
                  'version': '0.0.1',
                  'type': 'vanilla'
              })
      ])
  WINISO_IMAGE = t.WIN_ISO(
      "test",
      wib.ARCH_X86,
      "iso",
      uploads=[
          dest.Dest(
              cipd_src=sources.CIPDSrc(
                  package='experimental/mock/iso/test-1',
                  refs='latest',
                  platform='windows-amd64',
              ),
              tags={
                  'version': '0.0.1',
                  'type': 'vanilla'
              })
      ])

  SYSTEM = t.VM_DRIVE(
      name='system',
      ip=None,
      op=[
          dest.Dest(
              gcs_src=sources.GCSSrc(
                  bucket='chrome-windows-images', source='WIN-OUT/system.img'))
      ],
  )

  AARCH64_VM = t.VM_CONFIG(name='WinArm', version='latest', drives=[SYSTEM])

  WIN_IMAGE = t.WIN_IMAGE(
      'test', wib.ARCH_AARCH64, 'win', vm_config=AARCH64_VM, action_list=[])

  yield (api.test('Test WINPE trim dest', api.platform('win', 64)) +
         api.properties(WINPE_IMAGE) +
         # assert that the recipe execution was a success
         api.post_process(StatusSuccess) + api.post_process(DropExpectation))

  yield (api.test('Test WINISO trim dest') + api.properties(WINISO_IMAGE) +
         # assert that the recipe execution was a success
         api.post_process(StatusSuccess) + api.post_process(DropExpectation))

  yield (api.test('Test WIN trim dest') + api.properties(WIN_IMAGE) +
         # assert that the recipe execution was a success
         api.post_process(StatusSuccess) + api.post_process(DropExpectation))
