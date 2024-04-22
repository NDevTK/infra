# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PYTHON_VERSION_COMPATIBILITY = 'PY2+3'

DEPS = [
  'cloudkms',
  'recipe_engine/path',
]


def RunSteps(api):
  api.cloudkms.decrypt(
      'projects/PROJECT/locations/global/keyRings/KEYRING/cryptoKeys/KEY',
      api.path.start_dir / 'ciphertext',
      api.path.cleanup_dir / 'plaintext',
  )
  # Decrypt another file; the module shouldn't install cloudkms again.
  api.cloudkms.decrypt(
      'projects/PROJECT/locations/global/keyRings/KEYRING/cryptoKeys/KEY',
      api.path.start_dir / 'encrypted',
      api.path.cleanup_dir / 'decrypted',
  )

  api.cloudkms.sign(
      'projects/PROJECT/locations/LOCATION/keyRings/KEYRING/cryptoKeys/KEY',
      api.path.start_dir / 'chrome_build',
      api.path.start_dir / 'signed_bin',
  )
  #Sign another file; with service_account_json file not None
  api.cloudkms.sign(
      'projects/PROJECT/locations/LOCATION/keyRings/KEYRING/cryptoKeys/KEY',
      api.path.start_dir / 'build', api.path.start_dir / 'bin', 'service_acc')

  api.cloudkms.verify(
      'projects/PROJECT/locations/LOCATION/keyRings/KEYRING/cryptoKeys/KEY',
      api.path.start_dir / 'signed_chrome',
      api.path.start_dir / 'signature',
      api.path.cleanup_dir / 'result',
  )
  #Sign another file; with service_account_json file not None
  api.cloudkms.verify(
      'projects/PROJECT/locations/LOCATION/keyRings/KEYRING/cryptoKeys/KEY',
      api.path.start_dir / 'signed', api.path.start_dir / 'sign',
      api.path.cleanup_dir / 'status', 'service_acc')


def GenTests(api):
  yield api.test('simple')
