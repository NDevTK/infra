create {
  source {
    url {
      download_url: "https://url"
      version: "1.2.3"
    }
  }
  build {}
}

create {
  platform_re: "linux-arm64"
  build {
    tool: "tests/step_cross"
  }
}

upload { pkg_prefix: "tests" }
