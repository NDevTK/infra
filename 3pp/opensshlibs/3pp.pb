# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "https://cdn.openbsd.org/pub/OpenBSD/OpenSSH/portable/openssh-9.7p1.tar.gz"
      version: "9.7.1"
    }
    unpack_archive: true
  }

  build {
    tool: "tools/sed"
    dep: "static_libs/openssl"
    dep: "static_libs/zlib"
  }
}

upload { pkg_prefix: "static_libs" }
