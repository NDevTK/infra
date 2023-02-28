create {
  platform_re: "windows-amd64"
  source {
    url {
      download_url: "https://strawberryperl.com/download/5.32.1.1/strawberry-perl-5.32.1.1-64bit-portable.zip"
      version: "5.32.1.1"
      extension: ".zip"
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:perl:perl"
  }

  build {}
}

create {
  platform_re: "windows-386"
  source {
    url {
      download_url: "https://strawberryperl.com/download/5.32.1.1/strawberry-perl-5.32.1.1-32bit-portable.zip"
      version: "5.32.1.1"
      extension: ".zip"
    }
    unpack_archive: true
  }

  build {}
}

upload { pkg_prefix: "tools" }
