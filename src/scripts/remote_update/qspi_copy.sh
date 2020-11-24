#!/bin/sh

tar --no-same-owner -xvf /tmp/yocto.tar -C /tmp/ || exit 1
cd /tmp || exit 1
flashcp -v boot.bin /dev/mtd0 || exit 1
flashcp -v uImage /dev/mtd1 || exit 1
flashcp -v devicetree.dtb /dev/mtd2 || exit 1
flashcp -v uramdisk /dev/mtd5 || exit 1
