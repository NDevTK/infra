create {
  platform_re: "linux-.*"
  source {
    url {
      download_url: "https://github.com/westes/flex/releases/download/v2.6.4/flex-2.6.4.tar.gz"
      version: "2.6.4"
    }
    unpack_archive: true
    patch_version: "chromium.1"
  }

  build {
    tool: "tools/gettext"
    tool: "tools/help2man"
  }
}

upload { pkg_prefix: "tools" }