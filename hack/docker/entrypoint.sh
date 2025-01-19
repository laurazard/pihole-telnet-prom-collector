#!/usr/bin/env sh

# TODO(laurazard): improve/remove this
ssh -tt -o StrictHostKeyChecking=no -L 4711:127.0.0.1:4711 "root@${PI_COL_HOST}" </dev/null &
./pi-collector "$@"
