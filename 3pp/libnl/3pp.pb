create {
  platform_re: "linux-.*"
  source {
    url {
      download_url: "https://www.infradead.org/~tgr/libnl/files/libnl-3.2.25.tar.gz"
      version: "3.2.25"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:libnl_project:libnl"
    patch_version: "chromium.1"
  }

  build {
    tool: "tools/flex"
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }
