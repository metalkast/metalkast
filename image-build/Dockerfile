ARG UBUNTU_VERSION
ARG UBUNTU_RELEASE

FROM alpine AS image-download
ARG UBUNTU_VERSION
ARG UBUNTU_RELEASE
RUN wget "https://cloud-images.ubuntu.com/releases/${UBUNTU_VERSION}/release-${UBUNTU_RELEASE}/ubuntu-${UBUNTU_VERSION}-server-cloudimg-amd64.img" -O /ubuntu.img
# TODO: check shasum

FROM alpine AS iso-download
ARG UBUNTU_VERSION
ARG UBUNTU_DESKTOP_PATCH_RELEASE=2
RUN wget "https://releases.ubuntu.com/${UBUNTU_VERSION}.${UBUNTU_DESKTOP_PATCH_RELEASE}/ubuntu-${UBUNTU_VERSION}.${UBUNTU_DESKTOP_PATCH_RELEASE}-desktop-amd64.iso" -O live.iso || \
    wget "https://old-releases.ubuntu.com/releases/${UBUNTU_VERSION}/ubuntu-${UBUNTU_VERSION}.${UBUNTU_DESKTOP_PATCH_RELEASE}-desktop-amd64.iso" -O live.iso
# TODO: check shasum

FROM ubuntu:$UBUNTU_VERSION

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive \
    apt-get install -y \
        apparmor \
        apparmor-utils \
        binwalk \
        bridge-utils \
        casper \
        curl \
        genisoimage \
        git \
        libguestfs-tools \
        libvirt-clients \
        libvirt-daemon-system \
        linux-image-generic \
        live-boot \
        live-boot-initramfs-tools \
        make \
        p7zip-full \
        p7zip-rar \
        qemu-kvm \
        rsync \
        squashfs-tools \
        virt-manager

COPY --from=image-download /ubuntu.img /virt-customize/
COPY --from=iso-download /live.iso /virt-customize/

COPY virt-customize /virt-customize
WORKDIR /virt-customize

ARG UBUNTU_VERSION
ENV UBUNTU_VERSION=$UBUNTU_VERSION

ARG UBUNTU_RELEASE
ENV UBUNTU_RELEASE=$UBUNTU_RELEASE

# stabilizes printenv output
ENV HOSTNAME=image-build

COPY bootstrap-build.sh /virt-customize
RUN date > builder-timestamp.txt
ENTRYPOINT ["./build.sh"]
