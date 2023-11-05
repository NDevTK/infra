create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://github.com/NixOS/patchelf/releases/download/0.18.0/patchelf-0.18.0.tar.bz2"
      version: "0.18.0"
    }
    unpack_archive: true
  }
  build {}
}

upload { pkg_prefix: "tools" }
