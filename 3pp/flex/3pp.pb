create {
  platform_re: "linux-.*"
  source {
    url {
      download_url: "https://github.com/westes/flex/releases/download/v2.6.4/flex-2.6.4.tar.gz"
      version: "2.6.4"
    }
    unpack_archive: true
    patch_version: "chromium.3"
    cpe_base_address: "cpe:/a:flex_project:flex"
  }

  build {
    tool: "tools/gettext"
    tool: "tools/help2man"
    tool: "tools/sed"
  }

  package {
    version_file: ".versions/flex.cipd_version"
  }
}

upload { pkg_prefix: "tools" }
