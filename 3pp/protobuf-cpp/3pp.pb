create {
  platform_re: "linux-.*"
  source {
    script { name: "fetch.py" }
    unpack_archive: true
    patch_version: "chromium.2"
    cpe_base_address: "cpe:/a:protobuf_project:protobuf"
  }

  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }
