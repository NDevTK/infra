create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://people.redhat.com/sgrubb/libcap-ng/libcap-ng-0.8.3.tar.gz"
      version: "0.8.3"
    }
    unpack_archive: true
  }
  build {
    tool: "tools/sed"
  }
}

upload {
  pkg_prefix: "static_libs"
}
