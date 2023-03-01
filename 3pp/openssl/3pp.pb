create {
  source {
    url {
      download_url: "https://www.openssl.org/source/openssl-1.1.1t.tar.gz"
      version: "1.1.1t"
    }
    patch_version: "chromium.2"
    patch_dir: "patches"
    unpack_archive: true
    cpe_base_address: "cpe:/a:openssl_project:openssl"
  }
  build {}
}

create {
  platform_re: "windows-.*"
  build {
    tool: "tools/perl"
    tool: "tools/nasm"
  }
}

upload { pkg_prefix: "static_libs" }
