create {
  source {
    git {
      repo: "https://github.com/pypa/virtualenv.git"
      version_restriction {
        op: EQ
        val: "16.7.12"
      }
    }
    patch_dir: "patches"
    patch_version: "chromium.7"
    cpe_base_address: "cpe:/a:virtualenv:virtualenv"
  }
  build {
    dep: "build_support/pip_bootstrap"
  }
}

upload {
  pkg_prefix: "tools"
  universal: true
}
