create {
  platform_re: "windows-(amd64|arm64|386)"
  source {
    url {
      download_url: "https://go.microsoft.com/fwlink/?linkid=2196127"
      version: "10.1.22621.1"
      extension: ".exe"
    }
  }
}

upload { pkg_prefix: "tools" }
