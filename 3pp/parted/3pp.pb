create {
  platform_re: "linux-amd64"
   source {
    git {
      repo: "https://git.savannah.gnu.org/git/parted.git"
      tag_pattern: "v%s"
    }
  }
  build {
    tool: "tools/rsync"
    tool: "tools/gettext"
    tool: "tools/gperf"
    dep: "static_libs/libuuid"
  }
}

upload { pkg_prefix: "tools" }

