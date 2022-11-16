create {
  platform_re: "linux-amd64|mac-.*"
  source {
    script { name: "fetch.py" }
    unpack_archive: false
    patch_version: "chromium.2"
  }
  build {
    tool: "tools/autoconf"
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "tools" }
