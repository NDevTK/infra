create {
  platform_re: "linux-amd64"
  source {
    script{ name: "fetch.py" }
    unpack_archive: true
  }
}

upload {
  pkg_prefix: "tools"
}

