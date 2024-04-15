create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://github.com/github/codeql-action/releases/download/codeql-bundle-v2.17.0/codeql-bundle-linux64.tar.gz",
      version: "2.17.0"
    }
    unpack_archive: true
  }
}

upload { pkg_prefix: "tools" }
