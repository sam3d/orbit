#!/bin/bash
docker service create \
  --name edge \
  --network orbit \
  --mount type=bind,src=/var/run/orbit.sock,target=/var/run/orbit.sock \
  -p 80:80 -p 443:443 \
  orbit.sh/edge
