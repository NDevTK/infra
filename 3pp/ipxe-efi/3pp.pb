create {
  platform_re: "linux-amd64|linux-arm64"
  source {
    git {
      repo: "https://github.com/ipxe/ipxe.git"
      tag_pattern: "v%s"
    }
  }
  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }
