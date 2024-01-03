create {
  platform_re: "linux-amd64"
  source {
    git {
      repo: "https://gitlab.freedesktop.org/pixman/pixman"
      tag_pattern: "pixman-%s"
      version_restriction: { op: EQ val: "0.40.0" }
    }
    unpack_archive: true
    cpe_base_address: "cpe:/a:pixman:pixman"
    patch_version: "chromium.2"
  }
  build {
    tool: "tools/sed"
  }
}
