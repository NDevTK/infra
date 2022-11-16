create {
  platform_re: "linux-amd64|mac-.*"
  source {
    url {
      download_url: "https://archive.apache.org/dist/apr/apr-util-1.6.1.tar.gz"
      version: "1.6.1"
    }
    unpack_archive: true
    patch_dir: "patches"
    patch_version: "chromium.2"
  }

  build {
    tool: "tools/sed"

    dep: "static_libs/apr"
    dep: "static_libs/libexpat"
  }
}

upload { pkg_prefix: "static_libs" }
