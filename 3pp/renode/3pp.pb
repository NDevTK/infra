create {
  platform_re: "linux-.*"
  source {
    url {
      download_url: "https://github.com/renode/renode/releases/download/v1.14.0/renode-1.14.0.linux-portable.tar.gz"
      version: "renode-1.14.0"
    }
    unpack_archive: true
  }
  build {}
}

upload { pkg_prefix: "tools" }
