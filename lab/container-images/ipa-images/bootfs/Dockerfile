ARG ALPINE_VERSION=3.19.0

FROM alpine:${ALPINE_VERSION} as minirootfs-download

WORKDIR /tmp

RUN apk update
RUN apk add \
	lz4 \
	py3-pip

ARG ALPINE_VERSION
RUN mkdir alpine-minirootfs
RUN wget https://dl-cdn.alpinelinux.org/alpine/v${ALPINE_VERSION%.*}/releases/x86_64/alpine-minirootfs-${ALPINE_VERSION}-x86_64.tar.gz
RUN tar -xvf alpine-minirootfs-${ALPINE_VERSION}-x86_64.tar.gz -C alpine-minirootfs

FROM scratch AS alpine-setup
COPY --from=minirootfs-download /tmp/alpine-minirootfs /

# Base packages
RUN apk add \
	busybox-openrc \
	busybox-mdev-openrc \
	busybox-extras-openrc \
	kbd \
	openrc

# Binaries used by ironic-python-agent during its run.
RUN apk add \
  efibootmgr \
  findmnt \
  lsblk \
  sgdisk \
  parted \
  util-linux \
  coreutils \
  hdparm \
  qemu-img \
  lshw

# Some ironic-python-agent dependencies. Not all dependencies are available in apk, but it is still better to install those that are.
# Also packages required to build other dependencies.
RUN apk add \
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
RUN apk add \
	gcc \
	python3-dev \
	musl-dev

ARG IPA_VERSION=9.8.0
RUN pip3 install ironic-python-agent==${IPA_VERSION} --break-system-packages

RUN apk del \
	gcc \
	python3-dev \
	musl-dev

# Allow to login from console as root.
RUN passwd -d root

COPY files /

RUN for s in devfs dmesg hwdrivers mdev; do rc-update add $s sysinit; done
RUN for s in hostname syslog; do rc-update add $s boot; done
RUN for s in dhcp-all-interfaces networking nameserver-online ironic-python-agent; do rc-update add $s default; done

FROM alpine:${ALPINE_VERSION} as alpine-minirootfs-package
COPY --from=alpine-setup / /alpine-minirootfs

RUN apk update
RUN mkdir linux-kernel
RUN apk fetch linux-lts
RUN tar -xvf linux-lts-*.apk -C linux-kernel
RUN mv -v linux-kernel/lib/modules alpine-minirootfs/lib/

RUN cd alpine-minirootfs; find . -print0 | cpio --null --create --verbose --owner root:root --format=newc | gzip -6 > /ironic-python-agent.initramfs

FROM alpine:${ALPINE_VERSION}
COPY --from=alpine-minirootfs-package /ironic-python-agent.initramfs /
COPY --from=alpine-minirootfs-package /linux-kernel/boot/vmlinuz-lts /ironic-python-agent.kernel
