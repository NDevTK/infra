create {
  source {
    url {}
  }
  build {
    dep: "tests/step_dep"
    tool: "tests/step_tool"
    tool: "tests/step_tool_pinned@2@1.2.3"
  }
}

upload { pkg_prefix: "tests" }
