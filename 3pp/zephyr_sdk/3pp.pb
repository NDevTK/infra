create {
  platform_re: "linux-.*64|mac-.*"
  source {
    script{ name: "fetch.py" }
    unpack_archive: true
  }
}

create {
  platform_re: "windows-amd64"
  source {
    script{ name: "fetch.py" }
  }
  build {
    install: "install_win.sh"
    tool: "tools/7z"
  }
}

upload {
  pkg_prefix: "tools"
}

