create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://github.com/plougher/squashfs-tools/archive/4.6.1.tar.gz"
      version: "4.6.1"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:phillip_lougher:squashfs:4.6.1"
    patch_dir: "patches"
    patch_version: "chromium.2"
  }
  build {
    dep: "static_libs/libzstd"
    dep: "static_libs/zlib"
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }
