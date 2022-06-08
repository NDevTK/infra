create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://download.qemu.org/qemu-7.0.0-rc4.tar.xz"
      version: "7.0.0-rc4"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:qemu:qemu"
  }
  build {
    dep: "pixman"
  }
}

upload { pkg_prefix: "tools" }
