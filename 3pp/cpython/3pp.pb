create {
  source {
    patch_version: "chromium.39"
    cpe_base_address: "cpe:/a:python:python"
  }
  verify {
    test: "python_test.py"
  }
}
create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "https://www.python.org/ftp/python/2.7.18/Python-2.7.18.tgz"
      version: "2.7.18"
      extension: ".tgz"
    }
    unpack_archive: true
    patch_dir: "patches"
  }
  build {
    tool: "build_support/pip_bootstrap"
    tool: "tools/autoconf"
    tool: "tools/sed"
  }
}
create {
  platform_re: "mac-.*"
  source {
    patch_dir: "patches"
    patch_dir: "mac_patches"
  }
  build {
    dep: "static_libs/bzip2"
    dep: "static_libs/ncurses"
    dep: "static_libs/openssl"
    dep: "static_libs/readline"
    dep: "static_libs/sqlite"
    dep: "static_libs/zlib"
  }
}
create {
  platform_re: "linux-.*"
  build {
    dep: "static_libs/bzip2"
    dep: "static_libs/ncurses"
    dep: "static_libs/openssl"
    dep: "static_libs/readline"
    dep: "static_libs/sqlite"
    dep: "static_libs/zlib"
    dep: "static_libs/nsl"
  }
}
create {
  platform_re: "windows-.*"
  source {
    script {
      name: "fetch.py"
    }
  }
  build {
    install: "install_win.sh"
    tool: "build_support/pip_bootstrap"
    tool: "tools/lessmsi"
  }
  verify {
    test: "python_test.py"
  }
}
upload {
  pkg_prefix: "tools"
}
