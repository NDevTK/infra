create {
  platform_re: "linux-.*"
  source {
    script { name: "fetch.py" }
    unpack_archive: true
  }
  build {}
}

upload { pkg_prefix: "tools" }
