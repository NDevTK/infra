create {
  verify { test: "python_test.py" }
  source { patch_version: "chromium.35" }
  package {
    version_file: ".versions/cpython3.cipd_version"
    disable_latest_ref: true
  }
}

create {
  platform_re: "linux-.*|mac-.*"
  source {
    git {
      repo: "https://chromium.googlesource.com/external/github.com/python/cpython"
      tag_pattern: "v%s",

      # Python 3.10.11 was the last 3.10.x release that will have a binary
      # installer available for Windows. Pin to that version on all platforms
      # for consistency.
      version_restriction: { op: LT val: "3.10.12"}
    }
    patch_dir: "patches"
  }
  build {
    # no binutils on mac since it includes some tools like 'ar' that we don't
    # actually want
    tool: "build_support/pip_bootstrap@2@pip20.3.4.setuptools44.1.1.wheel0.37.1.chromium4"
    tool: "tools/autoconf"
    tool: "tools/sed"
  }
}

create {
  platform_re: "mac-.*"
  build {
    dep: "static_libs/bzip2"
    dep: "static_libs/libuuid"
    dep: "static_libs/ncursesw"
    dep: "static_libs/openssl"
    dep: "static_libs/readline"
    dep: "static_libs/sqlite"
    dep: "static_libs/xzutils"
    dep: "static_libs/zlib"
  }
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

    tool: "build_support/pip_bootstrap@2@pip20.3.4.setuptools44.1.1.wheel0.37.1.chromium4"
    tool: "tools/autoconf"
    tool: "tools/binutils"
    tool: "tools/sed"
  }
}

create {
  platform_re: "linux-arm.*|linux-mips.*|linux-riscv.*"
  build {
    tool: "build_support/pip_bootstrap@2@pip20.3.4.setuptools44.1.1.wheel0.37.1.chromium4"
    tool: "tools/autoconf"
    tool: "tools/binutils"
    tool: "tools/sed"            # Used by python's makefiles

    tool: "tools/cpython310"
  }
}

create {
  platform_re: "windows-.*"
  source { script { name: "fetch.py" } }
  build {
    tool: "build_support/pip_bootstrap@2@pip20.3.4.setuptools44.1.1.wheel0.37.1.chromium4"
    tool: "tools/lessmsi"

    install: "install_win.sh"
    # Note that the patches in patches_win are applied by install.sh
    # after the .msi files are extracted, rather than by the 3pp recipe.
  }
  verify { test: "python_test.py" }
}

upload {
  pkg_prefix: "tools"
  pkg_name_override: "cpython3"
}
