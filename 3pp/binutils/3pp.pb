create {
  platform_re: ".*-arm.*|.*-mips.*"
  unsupported: true
  source {
    cpe_base_address: "cpe:/a:gnu:binutils:-"
  }
}

create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "https://ftp.gnu.org/gnu/binutils/binutils-2.31.tar.gz"
      version: "2.31"
    }
    unpack_archive: true
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/texinfo"
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }
