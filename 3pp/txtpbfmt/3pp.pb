# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  # We are currently building this package only for linux platform.
  platform_re: "linux-.*"

  source {
    script { name: "fetch.py" }
    unpack_archive: true
  }

  build { tool: "tools/go" }
}

upload { pkg_prefix: "tools" }
