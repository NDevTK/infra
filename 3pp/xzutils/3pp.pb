create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "https://ixpeering.dl.sourceforge.net/project/lzmautils/xz-5.2.4.tar.gz"
      version: "5.2.4"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:xz_project:xz"
    patch_version: "chromium.3"
  }
  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "static_libs" }
