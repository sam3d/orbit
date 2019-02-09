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
  curl -L https://git.io/n-install | bash -s -- -y lts # Install Node.js
  apt-get install --yes glusterfs-server # Install glusterfs

  # Install golang
  add-apt-repository --yes ppa:longsleep/golang-backports
  apt-get update
  apt-get install --yes golang-go

  # Install pm2
  npm install -g pm2
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
    -p 80:5000 \
    --restart always --detach \
    orbit/console:dev
}

cleanup() {
  apt-get update

  echo "==> The development environment is now setup!"
  echo "--> Please wait for a few minutes as docker spins up the development"
  echo "    containers. You will not see any output for a short while."
}

main "$@"
