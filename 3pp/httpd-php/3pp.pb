create {
  platform_re: "linux-amd64|mac-.*"
  source {
    script { name: "fetch.py" }
    unpack_archive: false
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/autoconf"
  }
}

upload { pkg_prefix: "tools" }
