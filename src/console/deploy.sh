#!/bin/bash
docker service create \
  --name console \
  --network orbit \
  --mount type=bind,src=/var/run/orbit.sock,target=/var/run/orbit.sock \
  -p 6500:5000 \
  orbit.sh/console
