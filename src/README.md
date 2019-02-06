# Orbit source packages

The main components used by Orbit are located in this directory. Each directory corresponds to a different and individually deployable component.

This behaves as a kind of [monorepo](https://gomonorepo.org/) pattern.

## The different directories

### Overview

- **`agent`** - The primary Orbit runtime located directly on the host. It directly interacts with the helper in order to securely route traffic and configuration updates over the cluster. A running instance is located on every node.

  - **Exposes**: `/var/run/orbit.sock`
  - _Consumes_: `/var/run/docker.sock`
  - _Consumes_: GlusterFS command-line API

- **`cli`** - The command line client to interface with the agent. This will interact directly with the Orbit HTTP API socket.

  - _Consumes_: `var/run/orbit.sock`

- **`console`** _(containerised)_ - The web dashboard client. This will interact with the Orbit HTTP API socket (in the same way the CLI does) which then communicates data over the helper process.

  - _Consumes_: `/var/run/orbit.sock`

- **`edge`** _(containerised)_ - The edge-routing load balancer. This is primarily exposed on `80` and `443` and has its configuration controlled by the `agent`.

  - _Consumes_: `/var/run/orbit.sock`

- **`helper`** _(containerised)_ - A simple proxy for requests that can route data all over the cluster as and when is needed. A containerised helper instance is located on every node.

- _Consumes_: `/var/run/orbit.sock`
