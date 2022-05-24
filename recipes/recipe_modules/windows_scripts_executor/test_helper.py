# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from PB.recipes.infra.windows_image_builder import (offline_winpe_customization
                                                    as winpe)
from PB.recipes.infra.windows_image_builder import actions
from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import sources

from recipe_engine.post_process import DropExpectation, StatusFailure
from recipe_engine.post_process import StatusSuccess, StepCommandRE

#    Step data mock methods. Use these to mock the step outputs

# _gcs_stat is the mock output of gsutil stat command
_gcs_stat = """
{}:
    Creation time:          Tue, 12 Oct 2021 00:32:06 GMT
    Update time:            Tue, 12 Oct 2021 00:32:06 GMT
    Storage class:          STANDARD
    Content-Length:         658955236
    Content-Type:           application/octet-stream
    Metadata:
        orig:               {}
    Hash (crc32c):          oaYUgQ==
    Hash (md5):             +W9+CqZbFtYTZrUrDPltMw==
    ETag:                   CJOHnM3Pw/MCEAE=
    Generation:             1633998726431635
    Metageneration:         1
"""


def NEST(*args):
  """ NEST generates nested names for steps """
  return '.'.join(args)


def NEST_CONFIG_STEP(image):
  """ generate config step name for nesting """
  return 'execute config {}'.format(image)


def NEST_WINPE_CUSTOMIZATION_STEP(customization):
  """ generate winpe customization step name for nesting """
  return 'offline winpe customization {}'.format(customization)


def NEST_WINPE_INIT_STEP(arch, customization):
  """ generate winpe init step nesting names """
  return 'Init WinPE image modification {}'.format(
      arch) + ' in [CLEANUP]\\{}\\workdir'.format(customization)


def NEST_WINPE_DEINIT_STEP():
  """ generate winpe deinit step nesting names """
  return 'Deinit WinPE image modification'


def NEST_PROCESS_CUST():
  """ generate process customization header"""
  return 'Process the customizations'


def NEST_PIN_SRCS(cust):
  """ generate Pin Src step nesting name """
  return 'Pin resources from {}'.format(cust)


def NEST_DOWNLOAD_ALL_SRC(cust):
  """ Download all available packages step name"""
  return 'Download resources for {}'.format(cust)


def NEST_UPLOAD_CUST_OUTPUT(cust):
  """ Upload all gcs artifacts step name"""
  return 'Upload the output of {}'.format(cust)


def json_res(api, success=True, err_msg='Failed step'):
  """ generate a api.json object to mock outputs """
  if success:
    return api.json.output({'results': {'Success': success,}})
  return api.json.output({
      'results': {
          'Success': success,
          'Command': 'powershell',
          'ErrorInfo': {
              'Message': err_msg,
          },
      }
  })


def MOCK_WPE_INIT_DEINIT_SUCCESS(api, key, arch, image, customization):
  """ mock all the winpe init and deinit steps """
  return GEN_WPE_MEDIA(api, arch, image, customization) + \
        MOUNT_WIM(api, arch, image, customization) + \
        UMOUNT_WIM(api, image, customization) + \
        DEINIT_WIM_ADD_CFG_TO_ROOT(api, key, image, customization) + \
        CHECK_UMOUNT_WIM(api, image, customization)


def MOCK_WPE_INIT_DEINIT_FAILURE(api, arch, image, customization):
  """ mock all the winpe init and deinit steps on an action failure """
  return  GEN_WPE_MEDIA(api, arch, image, customization) + \
         MOUNT_WIM(api, arch, image, customization) + \
         UMOUNT_WIM( api, image, customization) + \
         CHECK_UMOUNT_WIM(api, image, customization, save=False)


def MOCK_CUST_IMG_WPE_INIT_DEINIT_SUCCESS(api, key, arch, image, customization):
  """ mock all the winpe init and deinit steps """
  return  MOUNT_WIM(api, arch, image, customization) + \
          UMOUNT_WIM(api, image, customization) + \
          DEINIT_WIM_ADD_CFG_TO_ROOT(api, key, image, customization) + \
          CHECK_UMOUNT_WIM(api, image, customization)


def MOCK_CUST_OUTPUT(api, url, success=True):
  retcode = 1
  if success:
    retcode = 0  #pragma: nocover
  return api.step_data(
      'gsutil stat {}'.format(url),
      api.raw_io.stream_output(_gcs_stat.format(url, url)),
      retcode=retcode,
  )


def GEN_WPE_MEDIA(api, arch, image, customization, success=True):
  """ Mock winpe media generation step """
  return api.step_data(
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(customization),
          NEST_WINPE_INIT_STEP(arch, customization),
          'PowerShell> Gen WinPE media for {}'.format(arch)),
      stdout=json_res(api, success))


def MOUNT_WIM(api, arch, image, customization, success=True):
  """ mock mount winpe wim step """
  return api.step_data(
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(customization),
          NEST_WINPE_INIT_STEP(arch, customization),
          'PowerShell> Mount wim to [CLEANUP]\\{}\\workdir\\mount'.format(
              customization)),
      stdout=json_res(api, success))


def UMOUNT_WIM(api, image, customization, success=True):
  """ mock unmount winpe wim step """
  return api.step_data(
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(customization),
          NEST_WINPE_DEINIT_STEP(),
          'PowerShell> Unmount wim at [CLEANUP]\\{}\\workdir\\mount'.format(
              customization)),
      stdout=json_res(api, success))


def DEINIT_WIM_ADD_CFG_TO_ROOT(api, key, image, customization, success=True):
  """ mock add cfg to root step in wpe deinit """
  return api.step_data(
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(customization),
          NEST_WINPE_DEINIT_STEP(),
          'PowerShell> Add cfg [CLEANUP]\\configs\\{}.cfg'.format(key)),
      stdout=json_res(api, success))


def GIT_PIN_FILE(api, cust, refs, path, data):
  """ mock git pin file step """
  return api.step_data(
      NEST(
          NEST_PROCESS_CUST(),
          NEST_PIN_SRCS(cust),
          'gitiles log: ' + '{}/{}'.format(refs, path),
      ),
      api.gitiles.make_log_test_data(data),
  )


def GCS_PIN_FILE(api, cust, url, pin_url='', success=True):
  """ mock gcs pin file action"""
  retcode = 1
  if success:
    retcode = 0
  if not pin_url:
    pin_url = url
  return api.step_data(
      NEST(NEST_PROCESS_CUST(), NEST_PIN_SRCS(cust),
           'gsutil stat {}'.format(url)),
      api.raw_io.stream_output(_gcs_stat.format(url, pin_url)),
      retcode=retcode,
  )


def GCS_DOWNLOAD_FILE(api, cust, bucket, source, success=True):
  """ mock gcs download file action"""
  retcode = 1
  if success:
    retcode = 0
  return api.step_data(
      NEST(
          NEST_DOWNLOAD_ALL_SRC(cust),
          'gsutil download gs://{}/{}'.format(bucket, source)),
      retcode=retcode,
  )


def ADD_FILE(api, image, customization, url, success=True):
  """ mock add file to image step """
  return api.step_data(
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(customization),
          'PowerShell> Add file {}'.format(url)),
      stdout=json_res(api, success))


def INSTALL_FILE(api, name, image, customization, success=True):
  """ mock install file to image step """
  return api.step_data(
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(customization),
          'PowerShell> Install package {}'.format(name)),
      stdout=json_res(api, success))


def INSTALL_DRIVER(api, name, image, customization, success=True):
  """ mock install driver to image step """
  return api.step_data(
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(customization),
          'PowerShell> Install driver {}'.format(name)),
      stdout=json_res(api, success))


def EDIT_REGISTRY(api, name, image, customization, success=True):
  """ mock registry edit action step"""
  return api.step_data(
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(customization),
          'PowerShell> Edit Offline Registry Key Features and Property {}'
          .format(name)),
      stdout=json_res(api, success))



#    Assert methods to validate that a certain step was run


def CHECK_UMOUNT_WIM(api, image, customization, save=True):
  """
      Post check that the wim was unmounted with either save or discard
  """
  args = ['.*'] * 10  # ignore matching the first 10 terms
  # check the last option
  if not save:
    args.append('-Discard')
  else:
    args.append('-Save')
  return api.post_process(
      StepCommandRE,
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(customization),
          NEST_WINPE_DEINIT_STEP(),
          'PowerShell> Unmount wim at [CLEANUP]\\{}\\workdir\\mount'.format(
              customization)), args)


def CHECK_GCS_UPLOAD(api, image, cust, source, destination, orig=''):
  """
      Post check the upload to GCS
  """
  if not orig:
    orig = destination
  args = ['.*'] * 11
  args[7] = 'x-goog-meta-orig:{}'.format(orig)  # ensure the orig meta url
  args[9] = source  # ensure the correct local src
  args[10] = destination  # ensure upload to correct location
  return api.post_process(
      StepCommandRE,
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(cust),
          NEST_WINPE_DEINIT_STEP(), NEST_UPLOAD_CUST_OUTPUT(cust),
          'gsutil upload {}'.format(destination)), args)


def CHECK_INSTALL_CAB(api, image, customization, action, args=None):
  """
      Post check for installation
  """
  wild_card = ['.*'] * 11
  if args:
    wild_card.append(*args)
  return api.post_process(
      StepCommandRE,
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(customization),
          'PowerShell> Install package {}'.format(action)), wild_card)


def CHECK_INSTALL_DRIVER(api, image, customization, action, args=None):
  """
      Post check for installation
  """
  wild_card = ['.*'] * 11
  if args:
    wild_card.append(*args)
  return api.post_process(
      StepCommandRE,
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(customization),
          'PowerShell> Install driver {}'.format(action)), wild_card)


def CHECK_CIPD_UPLOAD(api, image, cust, dest):
  """
      Post check the upload to GCS
  """
  # Wildcard args for everything + tags
  args = ['.*'] * (14 + len(dest.tags) * 2)
  # ref arg check
  args[7] = dest.cipd_src.refs  # check for correct refs
  # tags added in reverse order
  idx = 9 + len(dest.tags) * 2
  for tag, value in dest.tags.items():
    args[idx] = '{}:{}'.format(tag, value)
    idx -= 2
  package = dest.cipd_src.package
  platform = dest.cipd_src.platform
  return api.post_process(
      StepCommandRE,
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(cust),
          NEST_WINPE_DEINIT_STEP(), NEST_UPLOAD_CUST_OUTPUT(cust),
          'create {}/{}'.format(package, platform)), args)


def CHECK_ADD_FILE(api, image, cust, url, dest):
  """
      Check add file step
  """
  wild_card = ['.*'] * 14
  wild_card[3] = 'robocopy'
  wild_card[11] = dest

  return api.post_process(
      StepCommandRE,
      NEST(
          NEST_CONFIG_STEP(image), NEST_WINPE_CUSTOMIZATION_STEP(cust),
          'PowerShell> Add file {}'.format(url)), wild_card)



#   Generate proto configs helper functions


def WPE_IMAGE(image,
              arch,
              customization,
              sub_customization,
              action_list,
              up_dests=None):
  """ generates a winpe customization image """
  return wib.Image(
      name=image,
      arch=arch,
      customizations=[
          wib.Customization(
              offline_winpe_customization=winpe.OfflineWinPECustomization(
                  name=customization,
                  image_dests=up_dests,
                  offline_customization=[
                      actions.OfflineAction(
                          name=sub_customization, actions=action_list)
                  ]))
      ])
