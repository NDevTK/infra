create {
  source {
    patch_version: "chromium.6"
    cpe_base_address: "cpe:/a:git_project:git"
  }
  verify {
    test: "git_test.py"
  }
}
create {
  platform_re: "linux-.*|mac-.*"
  source {
    git {
      repo: "https://chromium.googlesource.com/external/github.com/git/git"
      tag_pattern: "v%s"
    }
  }
  build {
    tool: "tools/autoconf"
    tool: "tools/gettext"
    tool: "tools/sed"
    dep: "static_libs/curl"
    dep: "static_libs/libexpat"
    dep: "static_libs/pcre2"
    dep: "static_libs/zlib"
  }
}
create {
  platform_re: "windows-.*"
  source {
    script {
      name: "fetch_win.py"
    }
  }
  build {
    install: "install_win.sh"
  }
  package {
    alter_version_re: "(.*)\\.windows\\.\\d*(.*)"
    alter_version_replace: "\\1\\2"
  }
}
upload {
  pkg_prefix: "tools"
}
