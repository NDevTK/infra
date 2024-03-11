create {
  platform_re: "linux-amd64|mac-.*"
  source {
    url {
      download_url: "https://ftp.gnu.org/gnu/texinfo/texinfo-6.5.tar.gz"
      version: "6.5"
    }
    unpack_archive: true
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }
