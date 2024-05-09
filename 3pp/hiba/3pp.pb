# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  platform_re: "linux-.*|mac-.*"
  source {
    git {
      repo: "https://github.com/google/hiba.git"
      tag_pattern: "v%s"
    }
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/autoconf"
    tool: "tools/automake"
    tool: "tools/libtool"
    tool: "tools/sed"
    dep: "static_libs/opensshlibs"
    dep: "static_libs/openssl"
    dep: "static_libs/zlib"
  }
}

upload { pkg_prefix: "tools" }
