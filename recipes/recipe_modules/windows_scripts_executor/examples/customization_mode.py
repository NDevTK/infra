# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import (windows_image_builder as
                                                    wib)
from recipe_engine.post_process import DropExpectation, StatusSuccess
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

image = 'add_windows_driver_test'
customization = 'add_windows_driver'
key = '4587561855fcc569fc485e4f74b693870fd0d61eea5f40e9cef2ec9821d240a7'
url = 'gs://chrome-gce-images/WIB-WIM/{}.zip'


def RunSteps(api, config):
  # Test to ensure sanity of force build mode.
  api.windows_scripts_executor.init()
  custs = api.windows_scripts_executor.init_customizations(config)
  custs = api.windows_scripts_executor.process_customizations(custs, {})
  # If the image has force_build set, then the cust should be output
  custs = api.windows_scripts_executor.filter_executable_customizations(custs)
  assert (len(custs) > 0)


def GenTests(api):
  # Normal mode. Build the image if it doesn't exist
  yield (api.test('Mode: Normal; needs_build', api.platform('win', 64)) +
         api.properties(
             t.WPE_IMAGE(image, wib.ARCH_X86, customization, 'test', [])) +
         # mock output check to show that image doesn't exist
         t.MOCK_CUST_OUTPUT(api, url.format(key), success=False) +
         # assert that the execution was a success
         api.post_process(StatusSuccess) + api.post_process(DropExpectation))

  # Normal mode. Doesn't build the image as it exists
  yield (api.test('Mode: Normal; built not required', api.platform('win', 64)) +
         api.properties(
             t.WPE_IMAGE(image, wib.ARCH_X86, customization, 'test', [])) +
         # mock output check to show that image exists
         t.MOCK_CUST_OUTPUT(api, url.format(key), success=True) +
         # assert that the execution was a failure
         api.expect_exception('AssertionError') +
         api.post_process(DropExpectation))

  # Force build. Image exists (or not), build will happen
  yield (api.test('Mode: Normal; force build', api.platform('win', 64)) +
         api.properties(
             t.WPE_IMAGE(
                 image,
                 wib.ARCH_X86,
                 customization,
                 'test', [],
                 mode=wib.CustomizationMode.CUST_FORCE_BUILD)) +
         # assert  that the execution was a success
         api.post_process(StatusSuccess) + api.post_process(DropExpectation))
