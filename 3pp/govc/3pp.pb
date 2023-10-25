# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  platform_re: "linux-amd64"
  source {
    git {
      repo: "https://github.com/vmware/govmomi.git"
      tag_pattern: "v%s"
    }
  }
  build { tool: "tools/go" }
}

upload { pkg_prefix: "tools" }
