create {
  platform_re: "windows-amd64"
  source {
    script {
      name: ["pull_current_intel_drivers.py"]
      use_fetch_checkout_workflow: true
    }
  }
  build {
    install: ["install.sh"]
  }
}