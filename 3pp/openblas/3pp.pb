create {
  platform_re: "mac-.*"
  source {
    git {
      repo: "https://github.com/OpenMathLib/OpenBLAS.git"
      tag_pattern: "v%s"
    }
    patch_version: "chromium.1"
  }
  build {}
}

upload { pkg_prefix: "static_libs" }
