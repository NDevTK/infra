create {
  platform_re: "linux-.*|mac-.*"
  source {
    url {
      download_url: "https://mirrors.ibiblio.org/gnu/ed/ed-1.4.tar.gz"
      version: "1.4"
    }
    unpack_archive: true
    patch_version: "chromium.1"
  }
  build {
    tool: "tools/texinfo"
    tool: "tools/sed"
  }

  package {
    version_file: ".versions/ed.version"
  }
}

upload { pkg_prefix: "tools" }
