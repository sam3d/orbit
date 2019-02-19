const createApp = (name, dir, script) => ({
  name,
  cwd: `/opt/orbit/src/${dir}`,
  script: "nodemon",
  args: `--exec "${script}" --signal SIGTERM -L -e go --watch /opt/orbit/src`
});

module.exports = {
  apps: [
    createApp("orbitd", "daemon", "go run main.go"),
    createApp("orbitd-builder", "daemon", "make install"),
    createApp("orbit", "cli", "make install")
  ]
};
