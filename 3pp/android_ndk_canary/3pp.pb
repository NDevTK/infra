create {
  platform_re: "linux-amd64"

  source {
    patch_version: "chromium.1"
    script { name: "fetch.py" }
    unpack_archive: true
  }

  build {
    install: "install.sh"
    no_toolchain: true
  }

  package {
    disable_latest_ref: true
  }
}

upload {
  pkg_prefix: "tools"
  pkg_name_override: "android_ndk"
}
