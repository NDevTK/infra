create {
  platform_re: "linux-.*"
  source {
    git {
      repo: "https://github.com/protocolbuffers/protobuf.git"
      tag_pattern: "v%s"
      tag_filter_re: "v[0-9.]*$"  # Ignore rc releases
    }
    cpe_base_address: "cpe:/a:protobuf_project:protobuf"
  }

  build {
    tool: "tools/cmake"
    tool: "tools/sed"
    dep: "static_libs/zlib"
  }
}

upload { pkg_prefix: "tools" }
