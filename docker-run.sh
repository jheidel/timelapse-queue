#!/bin/sh

# This file contains one possible configuration for running the timelapse
# container as a persistent docker daemon.

TIMELAPSE_DIR=/home/jeff/

docker run \
  -d \
  --restart unless-stopped \
  -p 8888:8080 \
  --mount type=bind,source=${TIMELAPSE_DIR?},target=/mnt/fsroot \
  --name timelapse-queue \
  timelapse-queue
