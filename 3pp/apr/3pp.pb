create {
  platform_re: "linux-amd64|mac-.*"
  source {
    url {
      download_url: "https://archive.apache.org/dist/apr/apr-1.6.5.tar.gz"
      version: "1.6.5"
    }
    unpack_archive: true
    patch_dir: "patches"
    patch_version: "chromium.3"
  }

  build {
    tool: "tools/sed"
  }
}

upload { pkg_prefix: "static_libs" }
