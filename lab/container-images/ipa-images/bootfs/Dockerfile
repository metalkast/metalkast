ARG ALPINE_VERSION=3.15.4

FROM alpine:${ALPINE_VERSION}

WORKDIR /tmp

RUN apk update
RUN apk add \
	lz4 \
	py3-pip

ARG ALPINE_VERSION
RUN mkdir alpine-minirootfs
RUN wget https://dl-cdn.alpinelinux.org/alpine/v${ALPINE_VERSION%.*}/releases/x86_64/alpine-minirootfs-${ALPINE_VERSION}-x86_64.tar.gz
RUN tar -xvf alpine-minirootfs-${ALPINE_VERSION}-x86_64.tar.gz -C alpine-minirootfs

# Base packages
RUN apk add --root alpine-minirootfs \
	busybox-initscripts \
	kbd \
	openrc

# Binaries used by ironic-python-agent during its run.
RUN apk add --root alpine-minirootfs \
  efibootmgr \
  findmnt \
  lsblk \
  sgdisk \
  parted \
  util-linux \
  coreutils \
  hdparm \
  qemu-img

# Some ironic-python-agent dependencies. Not all dependencies are available in apk, but it is still better to install those that are.
# Also packages required to build other dependencies.
RUN apk add --root alpine-minirootfs \
	py3-cryptography \
	py3-bcrypt \
	py3-dateutil \
	py3-dnspython \
	py3-fasteners \
	py3-fixtures \
	py3-greenlet \
	py3-iso8601 \
	py3-netaddr \
	py3-netifaces \
	py3-pip \
	py3-pbr \
	py3-psutil \
	py3-rfc3986 \
	py3-testtools \
	py3-tz \
	py3-udev \
	py3-werkzeug \
	py3-wrapt \
	py3-yaml \
	py3-zeroconf

# Packages required to build API dependencies.
RUN apk add --root alpine-minirootfs \
	gcc \
	python3-dev \
	musl-dev

# pip can't download packages when it runs in chroot for some reasons.
# TODO: try to insall everything here and then just copy it to the destination image.
RUN mkdir alpine-minirootfs/tmp/pip3
COPY requirements.txt alpine-minirootfs/tmp/pip3/
RUN pip3 download -r alpine-minirootfs/tmp/pip3/requirements.txt --no-deps -d alpine-minirootfs/tmp/pip3/
RUN chroot alpine-minirootfs pip3 install --isolated --no-index --find-links /tmp/pip3 -r /tmp/pip3/requirements.txt
RUN rm -rf alpine-minirootfs/tmp/pip3

RUN apk del --root alpine-minirootfs \
	gcc \
	python3-dev \
	musl-dev

RUN mkdir linux-kernel
RUN apk fetch linux-lts
RUN tar -xvf linux-lts-*.apk -C linux-kernel

RUN mv -v linux-kernel/lib/modules alpine-minirootfs/lib/

# Allow to login from console as root.
RUN chroot alpine-minirootfs passwd -d root

COPY files alpine-minirootfs/

RUN for s in devfs dmesg hwdrivers mdev; do chroot alpine-minirootfs rc-update add $s sysinit; done
RUN for s in hostname syslog; do chroot alpine-minirootfs rc-update add $s boot; done
RUN for s in dhcp-all-interfaces networking nameserver-online ironic-python-agent; do chroot alpine-minirootfs rc-update add $s default; done

RUN cd alpine-minirootfs; find . -print0 | cpio --null --create --verbose --owner root:root --format=newc | gzip -6 > ../ironic-python-agent.initramfs
