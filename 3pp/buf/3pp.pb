create {
  # We are currently building this package only for linux platform.
  platform_re: "linux-.*"
}

create {
  source {
    git {
      repo: "https://chromium.googlesource.com/external/github.com/bufbuild/buf"
      tag_pattern: "upstream/v%s"

      # TODO: Remove this after we fix the symbolic link issue on windows.
      version_restriction: { op: EQ val: "1.9.0"}
    }
  }

  build { tool: "tools/go" }
}

upload { pkg_prefix: "go/github.com/bufbuild" }
