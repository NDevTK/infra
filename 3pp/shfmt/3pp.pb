create {
  # We are currently building this package only for linux platform.
  platform_re: "linux-.*"

  source {
    git {
      repo: "https://github.com/mvdan/sh"
      tag_pattern: "v%s"
    }
  }

  build { tool: "tools/go" }
}

upload { pkg_prefix: "tools" }
