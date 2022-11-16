create {
  platform_re: "linux-.*|mac-.*",
  source {
    url {
      download_url: "https://ftp.gnu.org/gnu/help2man/help2man-1.47.8.tar.xz"
      version: "1.47.8"
      extension: ".tar.xz"
    }
    unpack_archive: true
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }

