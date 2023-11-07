create {
  platform_re: "linux-.*"
  source {
    url {
      download_url: "https://github.com/libffi/libffi/archive/v3.2.1.tar.gz"
      version: "3.2.1"
    }
    patch_version: "chromium.4"
    unpack_archive: true
    cpe_base_address: "cpe:/a:libffi_project:libffi"
  }
  build {
    tool: "tools/autoconf"
    tool: "tools/automake"
    tool: "tools/libtool"
    tool: "tools/texinfo"
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "static_libs" }
