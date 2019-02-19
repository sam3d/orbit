#!/bin/bash
main() {
  ensure-environment
  install-deps
  cleanup
}

ensure-environment() {
  export DEBIAN_FRONTEND=noninteractive
  apt-get update
}

install-deps() {
  curl https://get.docker.com | bash # Install docker
  apt-get install --yes glusterfs-server # Install glusterfs

  # Install Node.js and nodemon
  curl -sL https://deb.nodesource.com/setup_10.x | bash -
  apt-get install --yes nodejs
  npm install -g nodemon

  # Install golang
  add-apt-repository --yes ppa:longsleep/golang-backports
  apt-get update
  apt-get install --yes golang-go
}

cleanup() {
  apt-get update
}

main "$@"
