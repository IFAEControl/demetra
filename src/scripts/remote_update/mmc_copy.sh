#!/bin/sh

mount /dev/mmcblk0p1 /mnt || exit
/bin/rm -f /mnt/* || exit
tar --no-same-owner -xvf /tmp/yocto.tar -C /mnt/ || exit
umount /mnt || exit
