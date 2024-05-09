# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  platform_re: "linux-.*|mac-.*"
    source {
    url {
      download_url: "https://media.luffy.cx/files/lldpd/lldpd-1.0.18.tar.gz"
      version: "1.0.18"
    }
    unpack_archive: true
  }

  build {
    tool: "tools/autoconf"
    tool: "tools/automake"
    tool: "tools/libtool"
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }
