create {
  # We are currently building this package only for linux platform.
  platform_re: "linux-.*"
}

create {
  source {
    git {
      repo: "https://github.com/bufbuild/buf.git"
      tag_pattern: "v%s"
    }
  }

  build { tool: "tools/go" }
}

upload { pkg_prefix: "go/github.com/bufbuild" }
