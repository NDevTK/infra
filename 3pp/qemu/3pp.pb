create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://download.qemu.org/qemu-8.2.2.tar.xz"
      version: "8.2.2"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:qemu:qemu"
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/sed"

    dep: "pixman"
    dep: "static_libs/attr"
    dep: "static_libs/libcap-ng"
  }
}

upload { pkg_prefix: "tools" }
