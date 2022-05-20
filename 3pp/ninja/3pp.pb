create {
  platform_re: ".*-386"
  unsupported: true
}

create {
  source {
    git {
      repo: "https://chromium.googlesource.com/external/github.com/ninja-build/ninja"
      tag_pattern: "v%s"
      # TODO(crbug.com/931218): migrate to CIPD 1.8.2 packages before version update.
      version_restriction: {
        op: EQ
        val: "1.8.2"
      }
    }
    patch_version: "chromium.1"
  }
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
  build {
    install: "install_linux-amd64.sh"
  }
}

upload { pkg_prefix: "tools" }
