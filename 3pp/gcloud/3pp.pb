create {
  platform_re: ".*-(amd64|arm64)"
  source {
    script { name: "fetch.py" }
    unpack_archive: true
    no_archive_prune: true
    patch_dir: "patches"
    patch_version: "chromium.2"
  }
  build {
    no_toolchain: true
    tool: "tools/cpython"
  }
}

upload { pkg_prefix: "tools" }
