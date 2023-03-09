# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  source {
    script {
      name: "../go/fetch.py"
      name: "prebuilt"
    }
    unpack_archive: true
    no_archive_prune: true
  }
  build {
    install: "install_official.sh"
  }
}

upload { pkg_prefix: "tools" }
