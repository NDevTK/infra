create {
  platform_re: "linux-.*|mac-.*"
  source {
    git {
      repo: "https://chromium.googlesource.com/external/github.com/swig/swig"
      tag_pattern: "rel-%s"
    }
    patch_dir: "patches"
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/autoconf"
    tool: "tools/automake"
    tool: "tools/sed"

    dep: "static_libs/pcre"
  }
}

upload { pkg_prefix: "tools" }
