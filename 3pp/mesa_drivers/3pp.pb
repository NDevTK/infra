create {
  platform_re: "linux-amd64"
  source {
    script {
      name: ["pull_current_mesa_drivers.py"]
    }
  }
}