# protobuf-cpp library pinned to version 21.x.
# Remove this once the nsjail package is able to compile with the
# latest protobuf.
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
  package {
    disable_latest_ref: true
  }
}

upload { pkg_prefix: "tools" }
