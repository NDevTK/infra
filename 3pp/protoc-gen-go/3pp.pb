create {
  # We are currently building this package only for linux platform.
  platform_re: "linux-.*"
}

create {
  source {
    git {
      repo: "https://chromium.googlesource.com/external/github.com/protocolbuffers/protobuf-go"
      tag_pattern: "v%s"
    }
    patch_version: "chromium.1"
  }

  build { tool: "tools/go" }
}

upload { pkg_prefix: "go/github.com/protocolbuffers" }
