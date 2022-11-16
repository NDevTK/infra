create {
  platform_re: "linux-.*|mac-.*"

  source {
    url {
      download_url: "https://curl.se/download/curl-7.59.0.tar.gz"
      version: "7.59.0"
    }
    patch_version: "chromium.4"
    unpack_archive: true
    cpe_base_address: "cpe:/a:curl_project:curl"
  }

  build {
    tool: "tools/sed"

    dep: "static_libs/libidn2"
    dep: "static_libs/zlib"
  }
}

create {
  platform_re: "linux-.*"

  build {
    tool: "tools/sed"

    dep: "static_libs/libidn2"
    dep: "static_libs/openssl"
    dep: "static_libs/zlib"
  }
}

upload { pkg_prefix: "static_libs" }
