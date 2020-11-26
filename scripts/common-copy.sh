#!/bin/bash 

DEST="$1"
SRC="$2"

# Only mercury-zx5 is supported 
MACHINE="$3"

BITSTREAM="$4"

#Explained in https://github.com/Xilinx/meta-xilinx/blob/master/README.booting.md#loading-via-sd

cp "$SRC/$MACHINE.dtb" "$DEST/devicetree.dtb"

# Generate boot.bin for enclustra
cp "$SRC/u-boot.elf" resources/binaries || exit 1
cp "$BITSTREAM" resources/binaries/fpga.bit || exit 1
(
  cd resources/binaries || exit
  mkbootimage boot.bif /tmp/boot.bin
)
cp /tmp/boot.bin "$SRC/boot.bin" || exit 1

# From microzed
#cp "$SRC/u-boot.img" "$DEST"

cp "$SRC/boot.bin" "$DEST" || exit 1
cp "$SRC/core-image-minimal-$MACHINE.cpio.gz.u-boot" "$DEST/uramdisk" || exit 1
cp "$SRC/uImage" "$DEST" || exit 1

# Convert bit to bin. bit format is not compatible
python resources/fpga-bit-to-bin.py --flip  "$BITSTREAM" "$DEST/fpga.bin"

cp resources/uEnv.txt "$DEST/uEnv.txt" || exit 1

# Check sizes
./scripts/check_image_files_size.py "$DEST" || exit
