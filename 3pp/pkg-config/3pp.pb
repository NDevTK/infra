# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  platform_re: "linux-.*|mac-.*"
  source {
    git {
      repo: "https://gitlab.freedesktop.org/pkg-config/pkg-config.git"
      tag_pattern: "pkg-config-%s"
    }
  }
  build {
    tool: "tools/autoconf"
  }
}

upload { pkg_prefix: "tools" }

    # tool: "tools/automake"
    # tool: "tools/libtool"


