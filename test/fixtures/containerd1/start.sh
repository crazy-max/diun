#!/bin/sh
set -eu

# Leave the delegated cgroup empty so runc can create task cgroups beneath it.
if [ -f /sys/fs/cgroup/cgroup.controllers ]; then
  mkdir -p /sys/fs/cgroup/init
  echo $$ > /sys/fs/cgroup/init/cgroup.procs
  sed 's/ / +/g; s/^/+/' /sys/fs/cgroup/cgroup.controllers > /sys/fs/cgroup/cgroup.subtree_control
fi

containerd &
daemon=$!
trap 'kill "$daemon"; wait "$daemon"' TERM INT EXIT
for attempt in $(seq 1 30); do
  if ctr version >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
ctr --namespace diun-e2e images pull --local --snapshotter native docker.io/library/busybox:1.36.1
ctr --namespace diun-e2e run --snapshotter native --detach --null-io --label diun.enable=true \
  docker.io/library/busybox:1.36.1 watched sleep 86400
wait "$daemon"
