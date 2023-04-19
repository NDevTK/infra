create {
  platform_re: "windows-.*"
  source {
    script {
      name: ["pull_current_intel_drivers.py"]
    }
  }
  build {
    install: ["install.sh"]
  }
  verify {
      test: [
          "pull_current_intel_drivers.py",
          "verify"
      ]
  }
}