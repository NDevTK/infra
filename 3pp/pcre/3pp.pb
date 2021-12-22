create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "https://ftp.pcre.org/pub/pcre/pcre-8.41.tar.gz"
      version: "8.41"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:pcre:pcre"
  }
  build {
  }
}
upload {
  pkg_prefix: "static_libs"
}
