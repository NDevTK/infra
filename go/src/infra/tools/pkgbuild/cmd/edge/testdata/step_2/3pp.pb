create {
  source {
    url {}
  }
  build {
    tool: "tests/step_tool_pinned@2@1.2.3"
  }
}

upload { pkg_prefix: "tests" }
