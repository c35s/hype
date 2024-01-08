#!/bin/sh

# This script becomes /init in .build/initrd.cpio.gz.
# It sets up a few basic mounts and runs /bin/sh.

set -eux

mount -t proc proc /proc
mount -t devtmpfs devtmps /dev

mkdir /dev/pts
mount -t devpts devpts /dev/pts

mount -t sysfs sysfs /sys
mount -t tmpfs tmpfs /tmp

hostname hype

exec /sbin/getty -ni -l /bin/sh 38400 hvc0
