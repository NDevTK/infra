create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "https://gitlab.com/bzip2/bzip2/-/archive/bzip2-1.0.6/bzip2-bzip2-1.0.6.tar.gz"
      version: "1.0.6"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:bzip:bzip2"
    patch_version: "chromium.1"
  }
  build {}
}

upload { pkg_prefix: "static_libs" }
