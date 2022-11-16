create {
  platform_re: "linux-amd64|mac-.*"
  source {
    url {
      download_url: "https://mirrors.kernel.org/gnu/bc/bc-1.07.1.tar.gz"
      version: "1.07.1"
    }
    unpack_archive: true
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/ed"
    tool: "tools/texinfo"
    tool: "tools/sed"
  }

  package {
    version_file: ".versions/bc.version"
  }
}

upload { pkg_prefix: "tools" }
