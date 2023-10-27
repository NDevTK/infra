create {
  platform_re: "linux-amd64|linux-arm64"
  source {
    git {
      repo: "https://github.com/ipxe/ipxe.git"
      fixed_commit: "115707c0edebad65f87525fed583fef73880016d"
    }
  }
  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }
