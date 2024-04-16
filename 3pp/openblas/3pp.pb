create {
  platform_re: "mac-.*"
  source {
    git {
      repo: "https://github.com/OpenMathLib/OpenBLAS.git"
      tag_pattern: "v%s"
    }
    patch_version: "chromium.2"
  }
  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "static_libs" }
