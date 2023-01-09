# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import (offline_winpe_customization
                                                    as winpe)
from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import actions
from PB.recipes.infra.windows_image_builder import sources

from recipe_engine.post_process import DropExpectation, StatusFailure
from recipe_engine.post_process import StatusSuccess, StepCommandRE
from recipe_engine.post_process import StatusException
from RECIPE_MODULES.infra.windows_scripts_executor import test_helper as t

DEPS = [
    'depot_tools/gitiles',
    'windows_scripts_executor',
    'recipe_engine/path',
    'recipe_engine/properties',
    'recipe_engine/platform',
    'recipe_engine/json',
    'recipe_engine/raw_io',
]

PYTHON_VERSION_COMPATIBILITY = 'PY3'

PROPERTIES = wib.Image

image = 'git_src_test'
customization = 'add_file_from_git'
key = '69c31bffdba451b237e80ee933b3667718166beb353bdb7c321ed167c8b51ce7'
arch = 'x86'


def RunSteps(api, config):
  api.windows_scripts_executor.init()
  custs = api.windows_scripts_executor.init_customizations(config)
  custs = api.windows_scripts_executor.process_customizations(custs, {})
  api.windows_scripts_executor.download_all_packages(custs)
  api.windows_scripts_executor.execute_customizations(custs)


def GenTests(api):
  # actions for adding files
  ACTION_ADD_STARTNET = actions.Action(
      add_file=actions.AddFile(
          name='add_startnet_file',
          src=sources.Src(
              git_src=sources.GITSrc(
                  repo='chromium.dev',
                  ref='HEAD',
                  src='windows/artifacts/startnet.cmd'),),
          dst='Windows\\System32',
      ))

  STARTNET_URL = 'chromium.dev/+/ef70cb069518e6dc3ff24bfae7f195de5099c377/' +\
                 'windows/artifacts/startnet.cmd'

  ACTION_ADD_DISKPART = actions.Action(
      add_file=actions.AddFile(
          name='add_diskpart_file',
          src=sources.Src(
              git_src=sources.GITSrc(
                  repo='chromium.dev',
                  ref='HEAD',
                  src='windows/artifacts/diskpart.txt'),),
          dst='Windows\\System32',
      ))

  DISKPART_URL = 'chromium.dev/+/ef70cb069518e6dc3ff24bfae7f195de5099c377/' +\
                 'windows/artifacts/diskpart.txt'

  yield (api.test('Add git src in action', api.platform('win', 64)) +
         # run a config for adding startnet file to wim
         api.properties(
             t.WPE_IMAGE(image, wib.ARCH_X86, customization,
                         'add_startnet_file', [ACTION_ADD_STARTNET])) +
         # mock all the init and deinit steps
         t.MOCK_WPE_INIT_DEINIT_SUCCESS(api, key, arch, image, customization) +
         # mock pin of the git src
         t.GIT_PIN_FILE(api, customization, 'HEAD',
                        'windows/artifacts/startnet.cmd', 'HEAD') +
         # mock adding the file to wim
         t.ADD_FILE(api, image, customization, STARTNET_URL) +
         api.post_process(StatusSuccess) +  # recipe should pass
         api.post_process(DropExpectation))

  yield (api.test('Fail download git_src', api.platform('win', 64)) +
         # run a config for adding startnet file to wim
         api.properties(
             t.WPE_IMAGE(image, wib.ARCH_X86, customization,
                         'add_startnet_file', [ACTION_ADD_STARTNET])) +
         # mock pin of the git src
         t.GIT_PIN_FILE(api, customization, 'HEAD',
                        'windows/artifacts/startnet.cmd', 'HEAD') +
         t.GIT_DOWNLOAD_FILE(api, customization, 'chromium.dev',
                             'ef70cb069518e6dc3ff24bfae7f195de5099c377',
                             'windows/artifacts/startnet.cmd', False) +
         api.post_process(StatusException) +
         api.expect_exception('SourceException') +
         api.post_process(DropExpectation))

  yield (api.test('Fail pin git_src', api.platform('win', 64)) +
         # run a config for adding startnet file to wim
         api.properties(
             t.WPE_IMAGE(image, wib.ARCH_X86, customization,
                         'add_startnet_file', [ACTION_ADD_STARTNET])) +
         # mock pin of the git src
         t.GIT_PIN_FILE(api, customization, 'HEAD',
                        'windows/artifacts/startnet.cmd', 'HEAD', False) +
         api.post_process(StatusException) +
         api.expect_exception('SourceException') +
         api.post_process(DropExpectation))

  # Adding same git src in multiple actions should trigger only one fetch action
  yield (
      api.test('Add same git src in multiple actions', api.platform('win',
                                                                    64)) +
      # run a config for adding startnet file to wim
      api.properties(
          t.WPE_IMAGE(image, wib.ARCH_X86, customization, 'add_startnet_file',
                      [ACTION_ADD_STARTNET, ACTION_ADD_STARTNET])) +
      # mock all the init and deinit steps
      t.MOCK_WPE_INIT_DEINIT_SUCCESS(api, key, arch, image, customization) +
      # mock pin of the git src, should only happen once
      t.GIT_PIN_FILE(api, customization, 'HEAD',
                     'windows/artifacts/startnet.cmd', 'HEAD') +
      # mock adding the file to wim
      t.ADD_FILE(api, image, customization, STARTNET_URL) +
      # mock adding the file to wim
      t.ADD_FILE(api, image, customization, STARTNET_URL + ' (2)') +
      api.post_process(StatusSuccess) +  # recipe should pass
      api.post_process(DropExpectation))

  yield (api.test('Add multiple git src in action', api.platform('win', 64)) +
         # run a config for adding startnet and diskpart files
         api.properties(
             t.WPE_IMAGE(image, wib.ARCH_X86, customization, 'action-1',
                         [ACTION_ADD_STARTNET, ACTION_ADD_DISKPART])) +
         # mock all the init and deinit steps
         t.MOCK_WPE_INIT_DEINIT_SUCCESS(api, key, arch, image, customization) +
         # mock pin of the git src
         t.GIT_PIN_FILE(api, customization, 'HEAD',
                        'windows/artifacts/startnet.cmd', 'HEAD') +
         # mock pin of the git src
         t.GIT_PIN_FILE(api, customization, 'HEAD',
                        'windows/artifacts/diskpart.txt', 'HEAD') +
         # mock adding the file to wim
         t.ADD_FILE(api, image, customization, STARTNET_URL) +
         # mock adding the file to wim
         t.ADD_FILE(api, image, customization, DISKPART_URL) +
         api.post_process(StatusSuccess) +  # recipe should pass
         api.post_process(DropExpectation))
