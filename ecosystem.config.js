module.exports = {
  apps: [
    {
      name: "moniplan-api",
      script: "/usr/local/go/bin/go",
      args: "run main.go",
      watch: ["."],
      interpreter: "none",
      env: {
        GO111MODULE: "on"
      }
    }
  ]
}
