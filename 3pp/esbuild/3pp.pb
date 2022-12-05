# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

create {
  source {
    git {
      repo: "https://github.com/evanw/esbuild.git"
      tag_pattern: "v%s"

      # A symbolic link in newer version of esbuild prevents windows from
      # downloading cached source code. The destination doesn't exist before
      # build so we can't copy it to avoid linking.
      # See also: crbug/1395682
      version_restriction: { op: EQ val: "0.15.16"}
    }
    patch_version: "chromium.2"
  }
  build {
    tool: "tools/go"
  }
}

upload { pkg_prefix: "tools" }
