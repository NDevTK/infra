create {
  platform_re: "mac-.*"
  source {
    git {
      repo: "https://github.com/OpenMathLib/OpenBLAS.git"
      tag_pattern: "v%s"
    }
  }
  build {}
}

upload { pkg_prefix: "static_libs" }
