# Orbit source packages

The main components used by Orbit are located in this directory. Each directory corresponds to a different and individually deployable component.

This behaves as a kind of [monorepo](https://gomonorepo.org/) pattern.

## The different directories

### Overview

- **`agent`** - The primary Orbit runtime on the host.
- **`cli`** - The command line client.
- **`console`** - The containerised web client.
- **`edge`** - The containerised edge-routing load balancer.
- **`helper`** - The containerised proxy for requests to the agent.
