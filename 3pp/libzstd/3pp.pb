create {
  platform_re: "linux-.*|mac-.*"
  source {
    git {
      repo: "https://github.com/facebook/zstd.git"
      tag_pattern: "v%s"
    }
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "static_libs" }

