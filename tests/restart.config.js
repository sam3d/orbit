const createApp = (name, dir, script) => ({
  name,
  cwd: `/opt/orbit/src/${dir}`,
  script: "nodemon",
  args: `--exec "${script}" -L -e go --watch /opt/orbit/src`
});

module.exports = {
  apps: [
    createApp("orbitd", "daemon", "make install && go run main.go"),
    createApp("orbit", "cli", "make install")
  ]
};
