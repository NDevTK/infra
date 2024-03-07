create {
  platform_re: "linux-amd64|mac-.*"
  source {
    url {
      download_url: "https://ftp.gnu.org/gnu/autoconf/autoconf-2.71.tar.gz"
      version: "2.71"
    }
    unpack_archive: true
    patch_dir: "patches"
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }
