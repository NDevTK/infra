create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://www.cairographics.org/releases/pixman-0.40.0.tar.gz"
      version: "0.40.0"
    }
    unpack_archive: true
  }
  build {
  }
}
