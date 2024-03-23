# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  platform_re: "mac-*"
  source {
    url {
      download_url: "https://github.com/google/hiba/archive/refs/tags/v1.4.tar.gz"
      version: "1.4"
    }
    unpack_archive: true
  }
  build {
    tool: "tools/autoconf"
    tool: "tools/automake"
    tool: "tools/libtool"
    dep: "static_libs/openssh"
    dep: "static_libs/openssl"
  }
}

upload { pkg_prefix: "tools" }
