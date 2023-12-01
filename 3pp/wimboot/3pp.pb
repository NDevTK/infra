# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  platform_re: "linux-amd64"
  source {
    script { name: "fetch.py" }
    unpack_archive: false
  }
}

upload { pkg_prefix: "tools" }