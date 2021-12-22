create {
  platform_re: "linux-.*"
  source {
    url {
      download_url: "https://dl.antmicro.com/projects/renode/builds/renode-1.12.0+20210521git8ae7fdfc.linux-portable.tar.gz"
      version: "renode-1.12.0+20210521git8ae7fdfc"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:rencode_project:rencode"
  }
  build {
  }
}
upload {
  pkg_prefix: "tools"
}
