main() {
  ensure-environment
  install-deps
  cleanup
}

ensure-environment() {
  export DEBIAN_FRONTEND=noninteractive

  apt-get update --yes
  apt-get upgrade --yes
}

install-deps() {
  curl https://get.docker.com | bash # Install docker
  apt-get install --yes glusterfs-server # Install glusterfs
}

cleanup() {
  apt-get update --yes
  apt-get upgrade --yes
}

main "$@"
