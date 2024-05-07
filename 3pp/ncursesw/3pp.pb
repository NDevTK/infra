create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "https://ftp.gnu.org/gnu/ncurses/ncurses-6.0.tar.gz"
      version: "6.0"
    }
    unpack_archive: true
    patch_dir: "patches"
    cpe_base_address: "cpe:/a:ncurses_project:ncurses"
    patch_version: "chromium.2"
  }
  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "static_libs" }
