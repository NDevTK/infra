create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "https://zlib.net/fossils/zlib-1.2.12.tar.gz"
      version: "1.2.12"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:zlib:zlib"
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "static_libs" }
