create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "http://sqlite.org/2017/sqlite-autoconf-3190300.tar.gz"
      version: "3.19.3"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:sqlite:sqlite"
    patch_version: "chromium.2"
  }
  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "static_libs" }
