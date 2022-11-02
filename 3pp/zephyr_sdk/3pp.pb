create {
  platform_re: "linux-amd64"
  source {
    url {
      download_url: "https://github.com/zephyrproject-rtos/sdk-ng/releases/download/v0.15.1/zephyr-sdk-0.15.1_linux-x86_64.tar.gz"
      version: "0.15.1"
    }
    unpack_archive: true
  }
}

upload {
  pkg_prefix: "tools"
}

