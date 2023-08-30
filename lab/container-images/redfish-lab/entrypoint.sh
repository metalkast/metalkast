#!/usr/bin/env bash
set -eEuo pipefail

# Enable nested cgroups to make libvirt work properly inside container
# https://gitlab.com/libvirt/libvirt/-/issues/163#note_577189452
# https://stackoverflow.com/a/76469328
# cgroup v2: enable nesting
if [ -f /sys/fs/cgroup/cgroup.controllers ]; then
	# move the processes from the root group to the /init group,
	# otherwise writing subtree_control fails with EBUSY.
	# An error during moving non-existent process (i.e., "cat") is ignored.
	mkdir -p /sys/fs/cgroup/init
	xargs -rn1 < /sys/fs/cgroup/cgroup.procs > /sys/fs/cgroup/init/cgroup.procs || :
	# enable controllers
	sed -e 's/ / +/g' -e 's/^/+/' < /sys/fs/cgroup/cgroup.controllers \
		> /sys/fs/cgroup/cgroup.subtree_control
fi

exec /usr/bin/supervisord -c /etc/supervisord.conf
