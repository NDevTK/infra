create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "https://www.7-zip.org/a/7z2201-src.tar.xz"
      version: "22.01"
      extension: "tar.xz"
    }
    unpack_archive: true
  }
  build {}
}

create {
  platform_re: "windows-.*"
  source {
    url {
      download_url: "https://www.7-zip.org/a/7z2201-src.tar.xz"
      version: "22.01"
      extension: "tar.xz"
    }
    unpack_archive: true
  }
  build {
    install: "install_win.sh"
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }

