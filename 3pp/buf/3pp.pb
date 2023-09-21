create {
  # We are currently building this package only for linux and mac.
  platform_re: "(linux|mac)-.*"

  source {
    git {
      repo: "https://chromium.googlesource.com/external/github.com/bufbuild/buf"
      tag_pattern: "upstream/v%s"
    }
  }

  build { tool: "tools/go" }
}

upload { pkg_prefix: "go/github.com/bufbuild" }
