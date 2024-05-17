create {
  platform_re: "linux-.*"
  source {
    git {
      repo: "https://chromium.googlesource.com/external/github.com/python/cpython"
      tag_pattern: "v%s",

      # Pin to 3.11.x for now.
      version_restriction: { op: LT val: "3.12a0"}
    }
    patch_dir: "patches"
    patch_version: "chromium.36"
  }
  build {
    # no binutils on mac since it includes some tools like 'ar' that we don't
    # actually want
    tool: "build_support/pip_bootstrap"
    tool: "tools/autoconf"
    tool: "tools/sed"
  }
  verify { test: "python_test.py" }
  package { version_file: ".versions/cpython3.cipd_version" }
}


create {
  platform_re: "linux-.*"
  build {
    dep: "static_libs/bzip2"
    dep: "static_libs/libffi"
    dep: "static_libs/libuuid"
    dep: "static_libs/ncursesw"
    dep: "static_libs/openssl"
    dep: "static_libs/readline"
    dep: "static_libs/sqlite"
    dep: "static_libs/xzutils"
    dep: "static_libs/zlib"

    tool: "build_support/pip_bootstrap"
    tool: "tools/autoconf"
    tool: "tools/binutils"
    tool: "tools/sed"
  }
}

create {
  platform_re: "linux-arm.*|linux-mips.*|linux-riscv.*"
  build {
    tool: "build_support/pip_bootstrap"
    tool: "tools/autoconf"
    tool: "tools/binutils"
    tool: "tools/sed"            # Used by python's makefiles

    tool: "tools/cpython3"
  }
}

upload { pkg_prefix: "tools" }
