create {
  platform_re: "linux-amd64"
   source {
    git {
      repo: "https://github.com/WayneD/rsync"
      tag_pattern: "v%s"
    }
  }
  build: {}
}

upload { pkg_prefix: "tools" }
