main() {
  ensure-environment
  install-deps
  setup-orbit
  cleanup
}

ensure-environment() {
  export DEBIAN_FRONTEND=noninteractive
  apt-get update
}

install-deps() {
  curl https://get.docker.com | bash # Install docker
  apt-get install --yes glusterfs-server # Install glusterfs

  # Install go and watcher tools
  add-apt-repository --yes ppa:longsleep/golang-backports
  apt-get update
  apt-get install --yes iwatch golang-go
}

setup-orbit() {
  # Build the console development image
  docker build \
    -f /opt/orbit/tests/dockerfiles/console.dockerfile \
    -t orbit/console:dev \
    /opt/orbit/src/console

  # Run the console development image
  docker run \
    -v /var/run/orbit.sock:/var/run/orbit.sock \
    -v /opt/orbit/src/console:/app \
    -v /tmp/orbit/console/node_modules:/app/node_modules \
    -p 6500:6500 \
    --restart always --detach \
    orbit/console:dev
}

cleanup() {
  apt-get update
}

main "$@"
