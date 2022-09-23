create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://www.7-zip.org/a/7z2201-src.tar.xz"
      version: "22.01"
      extension: "tar.xz"
    }
    patch_dir: "patches"
    unpack_archive: true
  }
  build {
    install: "install_linux_amd64.sh"
  }
}


create {
  platform_re: "linux-.*"
  source {
    url {
      download_url: "https://www.7-zip.org/a/7z2201-src.tar.xz"
      version: "22.01"
      extension: "tar.xz"
    }
    patch_dir: "patches"
    unpack_archive: true
  }
  build {}
}

create {
  platform_re: "mac-.*"
  source {
    url {
      download_url: "https://www.7-zip.org/a/7z2201-src.tar.xz"
      version: "22.01"
      extension: "tar.xz"
    }
    patch_dir: "patches"
    unpack_archive: true
  }
  build {
    install: "install_mac.sh"
  }
}

create {
  platform_re: "windows-386"
  source {
    url {
      download_url: "https://www.7-zip.org/a/7z2201-src.tar.xz"
      version: "22.01"
      extension: "tar.xz"
    }
    unpack_archive: true
  }
  build {
    install: "install_win_32.sh"
  }
}

create {
  platform_re: "windows-amd64"
  source {
    url {
      download_url: "https://www.7-zip.org/a/7z2201-src.tar.xz"
      version: "22.01"
      extension: "tar.xz"
    }
    unpack_archive: true
  }
  build {
    install: "install_win_64.sh"
  }
}

upload { pkg_prefix: "tools" }
