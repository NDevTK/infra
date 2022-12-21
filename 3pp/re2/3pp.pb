create {
  platform_re: "linux-amd64"
  source {
    # This project uses tags of the form YYYY-MM-DD, which are not currently
    # supported for 3pp's version comparison.
    url {
      download_url: "https://github.com/google/re2/archive/refs/tags/2022-12-01.tar.gz"
      version: "2022-12-01"
    }
    unpack_archive: true
  }
  build {
    tool: "tools/cmake"
  }
}

upload { pkg_prefix: "static_libs" }
