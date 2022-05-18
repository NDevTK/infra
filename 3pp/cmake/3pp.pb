create {
  platform_re: ".*-386"
  unsupported: true
  source {
    cpe_base_address: "cpe:/a:cmake_project:cmake"
  }
}

create {
  platform_re: "linux-.*|mac-.*"
  source {
    git {
      repo: "https://chromium.googlesource.com/external/github.com/Kitware/CMake"
      tag_pattern: "v%s"
      # CMake includes tags for release candidates like v3.22.0-rc1. Filter
      # the tag list to released versions, so the builder does not get stuck on
      # a prereleased one.
      tag_filter_re: "v[0-9.]*$"
    }
  }

  build {
    tool: "build_support/cmake_bootstrap"
    tool: "tools/ninja"
  }
}

create {
  platform_re: "linux-amd64"
  build {
    # ninja is built without docker,
    # and cmake also should be built in the same environment.
    no_docker_env: true
  }
}

upload { pkg_prefix: "tools" }
