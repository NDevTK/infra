create {
  # We are currently building this package only for linux platform.
  platform_re: "linux-amd64|linux-arm64"

  source {
    git {
      repo: "https://github.com/google/nsjail"
      tag_pattern: "%s"

      # We would like to use a fixed version of nsjail so that we can keep
      # its config stable in our codebase. Fixed to 3.3 for now.
      version_restriction: { op: EQ val: "3.3"}
    }
    patch_dir: "patches"
    patch_version: "chromium.2"
  }

  build {
    tool: "tools/flex"
    # TODO: Switch this back to the latest protobuf once we are building
    # an nsjail version with https://github.com/google/nsjail/issues/225
    # resolved.
    tool: "tools/protoc@2@21.12"
    dep: "tools/protobuf-cpp-21"
    dep: "tools/libnl"
  }
}

upload { pkg_prefix: "tools" }