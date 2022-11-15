create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://people.redhat.com/sgrubb/libcap-ng/libcap-ng-0.8.3.tar.gz"
      version: "0.8.3"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:libcap-ng_project:libcap-ng"
  }
  build {
    tool: "tools/sed"
  }
}

upload {
  pkg_prefix: "static_libs"
}
