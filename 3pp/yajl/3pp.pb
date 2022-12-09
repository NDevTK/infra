create {
  platform_re: "linux-amd64"
  source {
    git {
      repo: "https://github.com/lloyd/yajl.git"
      tag_pattern: "%s"
    }
  }
  build {
    tool: "tools/cmake"
  }
}

upload { pkg_prefix: "static_libs" }

