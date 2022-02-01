create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "http://xmlsoft.org/download/libxml2-2.9.12.tar.gz"
      version: "2.9.12"
    }
    unpack_archive: true
    patch_version: "chromium.1"
  }

  build {}
}

upload { pkg_prefix: "static_libs" }
