create {
  platform_re: "windows-.*"
  source {
    url {
      download_url: "https://www.nasm.us/pub/nasm/releasebuilds/2.16.01/nasm-2.16.01.tar.gz"
      version: "2.16.01"
    }
    unpack_archive: true
    patch_dir: "patches"
    cpe_base_address: "cpe:/a:nasm:nasm"
  }
  build {}
}

upload { pkg_prefix: "tools" }

