#!/bin/bash
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

  # Install Node.js and process manager
  curl -sL https://deb.nodesource.com/setup_10.x | bash -
  apt-get install --yes nodejs
  npm install -g pm2 nodemon
  pm2 startup

  # Install golang
  add-apt-repository --yes ppa:longsleep/golang-backports
  apt-get update
  apt-get install --yes golang-go
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

  # Start building and listening for go files
  pm2 start /opt/orbit/tests/restart.config.js
  pm2 save
}

cleanup() {
  apt-get update

  echo ""
  echo "==> The development environment is now setup!"
  echo "--> Please wait for a few minutes as docker spins up the development"
  echo "    containers. You will not see any output for a short while."
}

main "$@"
