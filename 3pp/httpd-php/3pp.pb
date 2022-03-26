create {
  platform_re: "linux-amd64|mac-.*"
  source {
    script { name: "fetch.py" }
    unpack_archive: false
  }
  build {
    tool: "tools/autoconf"
  }
}

create {
  platform_re: "windows-.*"
  source {
    script { name: "fetch_win.py" }
    unpack_archive: true
  }
  build {
    install: "install_win.sh"
  }
}

upload { pkg_prefix: "tools" }
