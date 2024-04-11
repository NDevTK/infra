# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  platform_re: "linux-.*|mac-.*"
  source {
    git {
      repo: "https://github.com/lldpd/lldpd.git"
    }
  }
  build {
    tool: "tools/autoconf"
    tool: "tools/automake"
    tool: "tools/libtool"
    tool: "tools/pkg-config"
  }
}

upload { pkg_prefix: "tools" }
