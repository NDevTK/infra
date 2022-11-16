create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://www.cairographics.org/releases/pixman-0.40.0.tar.gz"
      version: "0.40.0"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:pixman:pixman"
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/sed"
  }
}
