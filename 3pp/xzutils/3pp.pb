create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "https://tukaani.org/xz/xz-5.2.4.tar.gz"
      version: "5.2.4"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:lutils_project:lutils"
  }
  build {}
}

upload { pkg_prefix: "static_libs" }
