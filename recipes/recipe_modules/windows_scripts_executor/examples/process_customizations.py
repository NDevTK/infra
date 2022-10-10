# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
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
  custs = api.windows_scripts_executor.process_customizations(custs, {})


def GenTests(api):
  # Test context resolution by referencing a local customization. This provides
  # two customizations with one referncing the other.
  image = 'process_customizations_test'
  arch = wib.ARCH_AMD64
  cust1 = 'cust_1'
  cust2 = 'cust_2'
  sub_c1 = 'subcustomization_1'
  sub_c2 = 'subcustomiaztion_2'
  REF_IMAGE = t.WPE_IMAGE(image, arch, cust1, sub_c1, action_list=[])
  IMAGE_REFRENCING_PREV_IMAGE = t.WPE_IMAGE(
      image,
      arch,
      cust2,
      sub_c2,
      action_list=[],
      image_src=sources.Src(local_src='process_customizations_test-cust_1'))
  # Add the second customization to the first image. Now there are two
  # customizations in the REF_IMAGE with second referencing the first one.
  REF_IMAGE.customizations.extend(IMAGE_REFRENCING_PREV_IMAGE.customizations)

  yield (api.test('Test process_customizations Happy Path',
                  api.platform('win', 64)) +
         # input image with install file action without any args
         api.properties(REF_IMAGE) + api.post_process(StatusSuccess) +
         api.post_process(DropExpectation))

  # Test context resolution cyclical dependency. We create two customizations
  # with the first one referencing the second and the second one referncing the
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
  REF_IMAGE.customizations.extend(IMAGE_REFRENCING_PREV_IMAGE.customizations)
  yield (api.test('Test process_customizations Cyclical dependency',
                  api.platform('win', 64)) +
         # input image with install file action without any args
         api.properties(REF_IMAGE) + api.post_process(StatusFailure) +
         api.post_process(DropExpectation))
