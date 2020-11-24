#!/usr/bin/python3

import sys
import os

def size_of(file_name):
    if len(sys.argv) > 1:
        file_name = sys.argv[1] + "/" + file_name
    return os.stat(file_name).st_size/1024/1024

err = 0

BOOT_SIZE = 0x700000/1024/1024 # 7.0 MB
KERN_SIZE = 0x500000/1024/1024 # 5.0 MB
ROOTFS_SIZE = 0x32c0000/1024/1024 # 50.75 MB
DTS_SIZE = 0x80000/1024/1024 # 0.5 MB

if size_of("boot.bin") > BOOT_SIZE:
    print("boot.bin size is too big")
    err += 1

if size_of("uImage") > KERN_SIZE:
    print("uImage size is too big")
    err += 1

if size_of("devicetree.dtb") > DTS_SIZE:
    print("devictree.dtb size is too big")
    err += 1

if size_of("uramdisk") > ROOTFS_SIZE:
    print("uramdisk is too big")
    err += 1

if err == 0:
    print("\nIMAGE FILE SIZES ARE CORRECT\n")

sys.exit(err)
