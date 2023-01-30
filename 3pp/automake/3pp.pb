create {
  platform_re: "linux-.*|mac-.*"

  source {
    url {
      download_url: "https://ftp.gnu.org/gnu/automake/automake-1.15.tar.gz"
      version: "1.15"
    }
    cpe_base_address: "cpe:/a:gnu:automake"
    unpack_archive: true
    patch_dir: "patches"
    patch_version: "chromium.4"
  }

  build {
    tool: "tools/autoconf"
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }
