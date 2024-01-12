#!/usr/bin/env bash
set -eEuo pipefail

# https://help.ubuntu.com/community/LiveCDCustomization

mkdir extracted
7z x live.iso -oextracted
rm -rf extracted/\[BOOT\]/ extracted/casper/filesystem.squashfs

# enable console
sed -i -E 's/^([[:space:]]*linux.+---)/\1 console=tty0 console=ttyS0,115200n8/g' extracted/boot/grub/grub.cfg
# show entire boot output
sed -i -E 's/^([[:space:]]*linux.+) quiet(.+)/\1\2/g' extracted/boot/grub/grub.cfg
# setup network on init
sed -i -E 's/^([[:space:]]*linux.+)( ---.+)/\1 ip=dhcp\2/g' extracted/boot/grub/grub.cfg
# # set timeout to 1s to boot faster
sed -i -E 's/(set timeout=)30/\11/g' extracted/boot/grub/grub.cfg

image_base=$(find ./output -name '*.img')
image_live=$image_base.live
cp $image_base $image_live

qemu-img resize $image_live +4G
virt-customize -v -x -a $image_live --commands-from-file commands-live

mkdir edit
virt-copy-out -a $image_live / edit
rm -rf $image_live

rm extracted/casper/vmlinuz
cp edit/boot/vmlinuz extracted/casper/vmlinuz
chmod 644 extracted/casper/vmlinuz

mkdir initrdmount
unmkinitramfs -v extracted/casper/initrd initrdmount

cp -R initrdmount/main/conf conf
mv conf initrdconf
cp -R initrdmount/main/scripts initrdconf/scripts

kernel_version=$(file -bL extracted/casper/vmlinuz | grep -o 'version [^ ]*' | cut -d ' ' -f 2)
cp -r edit/lib/modules/$kernel_version /lib/modules/
mkinitramfs -d initrdconf -o ninitrd $kernel_version
rm extracted/casper/initrd
mv ninitrd extracted/casper/initrd

chmod +w extracted/casper/filesystem.manifest
chroot edit dpkg-query -W --showformat='${Package} ${Version}\n' > extracted/casper/filesystem.manifest
cp extracted/casper/filesystem.manifest extracted/casper/filesystem.manifest-desktop
sed -i '/ubiquity/d' extracted/casper/filesystem.manifest-desktop
sed -i '/casper/d' extracted/casper/filesystem.manifest-desktop

rm -f extracted/casper/*.squashfs
rm -f extracted/casper/*.squashfs.gpg

mksquashfs edit extracted/casper/filesystem.squashfs -comp xz
printf $(du -sx --block-size=1 edit | cut -f1) > extracted/casper/filesystem.size

rm -rf extracted/pool

cd extracted
rm -f md5sum.txt
find -type f -print0 | xargs -0 md5sum | grep -v isolinux/boot.cat | tee md5sum.txt
cd ..

# Add initial options first
cat <<EOF >xorriso.conf
-as mkisofs \\
-r -J --joliet-long \\
-o $(echo $image_base | sed 's/\.img$/-live.iso/') \\
EOF
# Use xorriso do the magic of figuring out options used to create original iso, making sure
# to append backslash to each line as required.
xorriso -report_about warning -indev "live.iso" -report_system_area as_mkisofs |
    sed -e 's|$| \\|'>>xorriso.conf
# Tell xorriso the root directory for the iso being created
echo 'extracted' >>xorriso.conf

# Modify options in xorriso.conf as desired or use as-is
xorriso -options_from_file xorriso.conf
