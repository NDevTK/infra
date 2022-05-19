create {
  platform_re: ".*-386"
  unsupported: true
}

create {
  source { git {
    repo: "https://chromium.googlesource.com/external/github.com/ninja-build/ninja"
    tag_pattern: "v%s"
  }}
}

create {
  platform_re: "mac-.*|linux-amd64"
  build {
    tool: "tools/re2c"
  }
}

create {
  platform_re: "linux-arm.*|linux-mips.*"
  build {
    tool: "tools/ninja"  # Depend on the bootstrapped version when cross-compiling
    tool: "tools/re2c"
  }
}

create {
  platform_re: "windows-.*"
  build {
    install: "install_bootstrap.sh"
  }
}

create {
  platform_re: "mac-.*"
  build {
    install: "install_mac.sh"
  }
}

create {
  platform_re: "linux-amd64"
  # manylinux2014 is too old to run chromium's clang++.
  build {
    no_docker_env: true
    install: "install_linux-amd64.sh"
  }
}

upload { pkg_prefix: "tools" }
