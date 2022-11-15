create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://download.savannah.nongnu.org/releases/attr/attr-2.5.1.tar.xz"
      version: "2.5.1"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:attr_project:attr"
  }
  build {
    tool: "tools/sed"
  }
}

upload {
  pkg_prefix: "static_libs"
}
