# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  platform_re: "linux-amd64"
  source {
    script { name: "fetch.py" }
    unpack_archive: true
    patch_version: 'cr0'
  }
  build {}
}

upload {
  pkg_prefix: "chromium/third_party/catapult"
}
