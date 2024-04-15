create {
  platform_re: "linux-amd64|linux-arm.*"
  source {
    url {
      download_url: "https://github.com/libffi/libffi/archive/v3.2.1.tar.gz"
      version: "3.2.1"
    }
    patch_version: "chromium.5"
    patch_dir: "patches"
    unpack_archive: true
    cpe_base_address: "cpe:/a:libffi_project:libffi"
  }
}

create {
  platform_re: "linux-riscv.*"
  source {
    url {
      download_url: "https://github.com/libffi/libffi/archive/v3.4.4.tar.gz"
      version: "3.4.4"
    }
    patch_version: "chromium.1"
    patch_dir: "patches-3.4.4"
    unpack_archive: true
    cpe_base_address: "cpe:/a:libffi_project:libffi"
  }
}

create {
  platform_re: "linux-.*"
  build {
    tool: "tools/autoconf"
    tool: "tools/automake"
    tool: "tools/libtool"
    tool: "tools/texinfo"
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "static_libs" }
