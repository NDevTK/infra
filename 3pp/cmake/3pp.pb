create {
  platform_re: ".*-386"
  unsupported: true
}

create {
  source {
    git {
      repo: "https://chromium.googlesource.com/external/github.com/Kitware/CMake"
      tag_pattern: "v%s"
      # CMake includes tags for release candidates like v3.22.0-rc1. Filter
      # the tag list to released versions, so the builder does not get stuck on
      # a prereleased one.
      tag_filter_re: "v[0-9.]*$"
    }
    patch_dir: "patches"
    cpe_base_address: "cpe:/a:cmake_project:cmake"
    patch_version: "chromium.3"
  }

  build {
    tool: "build_support/cmake_bootstrap"
    tool: "tools/ninja"
  }

  package {
    version_file: ".versions/cmake.cipd_version"
  }
}

upload { pkg_prefix: "tools" }
