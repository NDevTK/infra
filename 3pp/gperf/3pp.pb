create {
  platform_re: "linux-amd64"
   source {
    git {
      repo: "https://github.com/rurban/gperf"
      tag_pattern: "v%s"
    }
  }
  build {
    tool: "tools/texinfo"
  }
}

upload { pkg_prefix: "tools" }

