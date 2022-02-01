create {
  source {
    script { name: "fetch.py" }
    unpack_archive: true
    cpe_base_address: "cpe:/a:psutil_project:psutil"
  }
}

upload {
  pkg_prefix: "tools"
  universal: true
}
