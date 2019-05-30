#!/bin/sh

# This file contains one possible configuration for running the timelapse
# container as a persistent docker daemon.

TIMELAPSE_DIR=/data/share/

docker run \
  -d \
  --restart unless-stopped \
  -p 8888:8080 \
  --mount type=bind,source=${TIMELAPSE_DIR?},target=/mnt/fsroot --user 1002:1002 \
  --cpus=5.5 --memory=20g --memory-swap=20g \
  --name timelapse-queue \
  jheidel/timelapse-queue
