# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  platform_re: "mac-.*"
  source {
    git {
      repo: "https://github.com/openssh/openssh-portable.git"
      tag_pattern: "V_%s"
      version_join: "_"
    }
  }

  build {
    tool: "tools/autoconf"
    tool: "tools/automake"
    tool: "tools/libtool"
    dep: "static_libs/openssl"
    dep: "static_libs/zlib"
  }
}

upload { pkg_prefix: "static_libs" }
