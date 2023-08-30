ARG ALPINE_VERSION=3.15.4

# Using Ubuntu grub as it is signed by trusted CA.
FROM ubuntu:20.04
WORKDIR /tmp

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update
RUN apt-get install -y \
  dosfstools \
  grub-efi-amd64-signed \
  mtools \
  shim-signed

RUN dd bs=1024 count=6400 if=/dev/zero of=esp.img
RUN mkfs.msdos -F 12 -n 'ESP_IMAGE' ./esp.img
RUN mmd   -i esp.img ::EFI
RUN mmd   -i esp.img ::EFI/BOOT
RUN mcopy -i esp.img -v /usr/lib/shim/shimx64.efi.signed ::EFI/BOOT/BOOTX64.EFI
RUN mcopy -i esp.img -v /usr/lib/grub/x86_64-efi-signed/grubx64.efi.signed ::EFI/BOOT/grubx64.efi

