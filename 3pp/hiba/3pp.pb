# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  platform_re: "mac-.*"
  source {
    git {
      repo: "https://github.com/google/hiba.git"
      tag_pattern: "v%s"
    }
  }
  build {
    tool: "tools/autoconf"
    tool: "tools/automake"
    tool: "tools/libtool"
    dep: "static_libs/opensshlibs"
    dep: "static_libs/openssl"
  }
}

upload { pkg_prefix: "tools" }
