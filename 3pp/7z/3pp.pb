create {
  platform_re: "linux-amd64|mac-.*|windows-.*"
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

upload { pkg_prefix: "tools" }
